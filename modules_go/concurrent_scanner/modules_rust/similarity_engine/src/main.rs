use serde::{Deserialize, Serialize};
use std::io::{self, Read};
use std::collections::HashSet;

#[derive(Debug, Deserialize)]
struct Identity {
    id: String,
    username: String,
    bio: String,
    avatar_hash: String,
    links: Vec<String>,
}

#[derive(Debug, Serialize)]
struct SimilarityResult {
    id_a: String,
    id_b: String,
    score: f64,
    reasons: Vec<String>,
}

fn main() -> io::Result<()> {
    // 1. Read input from stdin
    let mut input = String::new();
    io::stdin().read_to_string(&mut input)?;

    let identities: Vec<Identity> = serde_json::from_str(&input)
        .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))?;

    let mut results = Vec::new();

    // 2. Pairwise comparison
    for i in 0..identities.len() {
        for j in i + 1..identities.len() {
            let a = &identities[i];
            let b = &identities[j];

            let (score, reasons) = compare_identities(a, b);
            if score >= 0.65 {
                results.push(SimilarityResult {
                    id_a: a.id.clone(),
                    id_b: b.id.clone(),
                    score,
                    reasons,
                });
            }
        }
    }

    // 3. Output results to stdout
    let output = serde_json::to_string(&results)
        .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;
    println!("{}", output);

    Ok(())
}

fn compare_identities(a: &Identity, b: &Identity) -> (f64, Vec<String>) {
    let mut score = 0.0;
    let mut reasons = Vec::new();
    let mut has_strong_signal = false;

    // 1. Username Scoring (Levenshtein)
    let u_sim = levenshtein_sim(&a.username.to_lowercase(), &b.username.to_lowercase());
    if u_sim == 1.0 {
        score += 0.5;
        reasons.push("exact username match".to_string());
        has_strong_signal = true;
    } else if u_sim > 0.85 {
        score += 0.2;
        reasons.push("high-confidence fuzzy username match".to_string());
    } else {
        score -= 0.3;
    }

    // 2. Avatar Hash
    if !a.avatar_hash.is_empty() && a.avatar_hash == b.avatar_hash {
        score += 0.3;
        reasons.push("identical avatar signature".to_string());
        has_strong_signal = true;
    }

    // 3. Shared Links
    let set_a: HashSet<_> = a.links.iter().collect();
    let set_b: HashSet<_> = b.links.iter().collect();
    let common_links = set_a.intersection(&set_b).count();
    if common_links > 0 {
        score += 0.2;
        reasons.push("shared external links".to_string());
        has_strong_signal = true;
    }

    // 4. Bio (Jaccard)
    let j_sim = jaccard_sim(&a.bio, &b.bio);
    if j_sim > 0.4 {
        score += 0.15;
        reasons.push(format!("smart bio match (sim: {:.2})", j_sim));
    } else if !a.bio.is_empty() && !b.bio.is_empty() && j_sim == 0.0 {
        score -= 0.2;
    }

    if !has_strong_signal {
        return (0.0, Vec::new());
    }

    (score.clamp(0.0, 1.0), reasons)
}

fn levenshtein_sim(s1: &str, s2: &str) -> f64 {
    if s1 == s2 { return 1.0; }
    let len1 = s1.chars().count();
    let len2 = s2.chars().count();
    if len1 == 0 || len2 == 0 { return 0.0; }

    let mut matrix = vec![vec![0; len2 + 1]; len1 + 1];
    for i in 0..=len1 { matrix[i][0] = i; }
    for j in 0..=len2 { matrix[0][j] = j; }

    for (i, c1) in s1.chars().enumerate() {
        for (j, c2) in s2.chars().enumerate() {
            let cost = if c1 == c2 { 0 } else { 1 };
            matrix[i + 1][j + 1] = (matrix[i][j + 1] + 1)
                .min(matrix[i + 1][j] + 1)
                .min(matrix[i][j] + cost);
        }
    }

    let dist = matrix[len1][len2];
    1.0 - (dist as f64 / len1.max(len2) as f64)
}

fn jaccard_sim(s1: &str, s2: &str) -> f64 {
    let tokens1: HashSet<_> = s1.split_whitespace().map(|s| s.to_lowercase()).collect();
    let tokens2: HashSet<_> = s2.split_whitespace().map(|s| s.to_lowercase()).collect();
    
    if tokens1.is_empty() || tokens2.is_empty() { return 0.0; }
    
    let intersection = tokens1.intersection(&tokens2).count();
    let union = tokens1.union(&tokens2).count();
    
    intersection as f64 / union as f64
}
