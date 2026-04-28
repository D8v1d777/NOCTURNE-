import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity

class FeatureExtractor:
    """Extracts numerical features from identity pairs for ML scoring."""
    
    def __init__(self):
        # TF-IDF vectorizer for bio similarity.
        # It needs to be fitted on a corpus to be effective.
        # For now, we'll fit it on the two bios being compared.
        # In a production system, it would be pre-fitted on a large corpus of bios.
        self.vectorizer = TfidfVectorizer(stop_words='english')

    def _levenshtein_norm(self, s1, s2):
        """Normalized Levenshtein distance."""
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
        bio_sim = 0.0
        if bio_a and bio_b:
            try:
                # Fit and transform on the fly for simplicity in this example.
                # In production, vectorizer would be pre-fitted.
                tfidf = self.vectorizer.fit_transform([bio_a, bio_b])
                bio_sim = cosine_similarity(tfidf[0:1], tfidf[1:2])[0][0]
            except ValueError: # Handle empty vocabulary if bios are too short/similar
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
                             and identity_a.get("email") is not None and identity_a.get("email") != "") else 0.0
        
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

    def get_feature_names(self):
        """Returns the ordered list of feature names."""
        return [
            "username_similarity",
            "exact_username_match",
            "email_match",
            "bio_similarity",
            "avatar_similarity",
            "platform_overlap",
            "link_overlap"
        ]