import json
import numpy as np
from sklearn.linear_model import LogisticRegression
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity

class IdentityMLScorer:
    """Probabilistic scoring engine for identity correlation using Logistic Regression."""
    
    def __init__(self):
        self.model = LogisticRegression()
        self.vectorizer = TfidfVectorizer(stop_words='english')
        self._is_trained = False
        self._bootstrap_model()

    def _levenshtein_norm(self, s1, s2):
        """Normalized Levenshtein distance matching the Go core implementation."""
        if not s1 or not s2: return 0.0
        if s1 == s2: return 1.0
        
        rows = len(s1) + 1
        cols = len(s2) + 1
        dist = np.zeros((rows, cols), dtype=int)
        
        for i in range(1, rows): dist[i, 0] = i
        for i in range(1, cols): dist[0, i] = i
        
        for col in range(1, cols):
            for row in range(1, rows):
                cost = 0 if s1[row-1] == s2[col-1] else 1
                dist[row, col] = min(dist[row-1, col] + 1,      # deletion
                                     dist[row, col-1] + 1,      # insertion
                                     dist[row-1, col-1] + cost) # substitution
                                     
        return 1.0 - (dist[len(s1), len(s2)] / max(len(s1), len(s2)))

    def _hamming_norm(self, h1, h2):
        """Normalized Hamming distance for perceptual hashes."""
        if not h1 or not h2 or len(h1) != len(h2): return 0.0
        diff = sum(c1 != c2 for c1, c2 in zip(h1, h2))
        return 1.0 - (diff / len(h1))

    def _jaccard(self, list_a, list_b):
        """Jaccard similarity for link overlap."""
        set_a, set_b = set(list_a or []), set(list_b or [])
        if not set_a and not set_b: return 0.0
        intersection = len(set_a.intersection(set_b))
        union = len(set_a.union(set_b))
        return intersection / union

    def extract_features(self, identity_a, identity_b):
        """Transform identity pairs into a feature vector for ML evaluation."""
        # 1. Username
        u_a = identity_a.get("username", "").lower()
        u_b = identity_b.get("username", "").lower()
        u_sim = self._levenshtein_norm(u_a, u_b)
        
        # 2. Bio Similarity (TF-IDF Cosine)
        bio_a = identity_a.get("bio") or ""
        bio_b = identity_b.get("bio") or ""
        if bio_a and bio_b:
            tfidf = self.vectorizer.fit_transform([bio_a, bio_b])
            bio_sim = cosine_similarity(tfidf[0:1], tfidf[1:2])[0][0]
        else:
            bio_sim = 0.0
            
        # 3. Avatar
        h_a = identity_a.get("avatar_hash", "")
        h_b = identity_b.get("avatar_hash", "")
        avatar_sim = self._hamming_norm(h_a, h_b)
        
        # 4. Links
        link_sim = self._jaccard(identity_a.get("links"), identity_b.get("links"))
        
        # 5. Exact Matches
        exact_u = 1.0 if u_a == u_b and u_a != "" else 0.0
        email_match = 1.0 if (identity_a.get("email") == identity_b.get("email") 
                             and identity_a.get("email") is not None) else 0.0
        
        # 6. Platform context
        platform_overlap = 1.0 if identity_a.get("platform") == identity_b.get("platform") else 0.0

        return {
            "username_similarity": u_sim,
            "exact_username_match": exact_u,
            "email_match": email_match,
            "bio_similarity": bio_sim,
            "avatar_similarity": avatar_sim,
            "platform_overlap": platform_overlap,
            "link_overlap": link_sim
        }

    def _bootstrap_model(self):
        """Trains the Logistic Regression model using synthetic OSINT data samples."""
        X_train = []
        y_train = []

        # Positive Samples (Strong username + bio/links)
        X_train.append([1.0, 1.0, 0.0, 0.8, 0.0, 0.0, 0.5]) # Same user, high link overlap
        y_train.append(1)
        X_train.append([0.9, 0.0, 1.0, 0.4, 0.9, 0.0, 0.0]) # Same email/avatar, fuzzy user
        y_train.append(1)
        X_train.append([1.0, 1.0, 0.0, 0.1, 0.0, 0.0, 1.0]) # Exact links/user
        y_train.append(1)

        # Negative Samples (Random noise / collisions)
        X_train.append([0.2, 0.0, 0.0, 0.1, 0.0, 1.0, 0.0]) # Random people on same platform
        y_train.append(0)
        X_train.append([0.5, 0.0, 0.0, 0.0, 0.2, 0.0, 0.0]) # Weak fuzzy username match
        y_train.append(0)
        X_train.append([1.0, 1.0, 0.0, 0.0, 0.0, 1.0, 0.0]) # Same user, same platform (usually separate accounts)
        y_train.append(0)

        self.model.fit(np.array(X_train), np.array(y_train))
        self._is_trained = True

    def score_identity_match(self, identity_a, identity_b):
        """Evaluates two records and returns match probability and decision."""
        features_dict = self.extract_features(identity_a, identity_b)
        vector = np.array(list(features_dict.values())).reshape(1, -1)
        
        probability = self.model.predict_proba(vector)[0][1]
        
        return {
            "probability": round(float(probability), 4),
            "decision": probability >= 0.75,
            "features": features_dict
        }

    def batch_score(self, pairs):
        """Optimized batch processing for high-volume correlation tasks."""
        results = []
        for a, b in pairs:
            results.append(self.score_identity_match(a, b))
        return results

if __name__ == "__main__":
    # Example Scenarios
    scorer = IdentityMLScorer()
    
    # Scenario A: Highly likely match
    id1 = {
        "username": "shadow_coder", 
        "bio": "OSINT researcher and Go enthusiast",
        "links": ["https://shadow.io"],
        "avatar_hash": "a1b2c3d4"
    }
    id2 = {
        "username": "shadow_coder", 
        "bio": "I code in Go and explore digital shadows",
        "links": ["https://shadow.io"],
        "avatar_hash": "a1b2c3d4"
    }
    
    # Scenario B: Unlikely match
    id3 = {
        "username": "coder_123",
        "platform": "reddit"
    }

    print("--- SCENARIO A (MATCH) ---")
    print(json.dumps(scorer.score_identity_match(id1, id2), indent=2))
    
    print("\n--- SCENARIO B (NO MATCH) ---")
    print(json.dumps(scorer.score_identity_match(id1, id3), indent=2))