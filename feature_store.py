import sqlite3
import json
import hashlib

class FeatureStore:
    """
    Manages a SQLite database to cache feature vectors for identity pairs.
    This avoids recomputing expensive features.
    """
    def __init__(self, db_path):
        self.db_path = db_path
        self._create_table()

    def _create_table(self):
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS features (
                pair_hash TEXT PRIMARY KEY,
                features_json TEXT,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        """)
        conn.commit()
        conn.close()

    def _generate_pair_hash(self, identity_a, identity_b):
        """Generates a consistent hash for an identity pair, regardless of order."""
        # Sort IDs to ensure consistent hash for (A,B) vs (B,A)
        id1 = identity_a.get("id", "")
        id2 = identity_b.get("id", "")
        sorted_ids = tuple(sorted((id1, id2)))
        return hashlib.sha256(json.dumps(sorted_ids).encode()).hexdigest()

    def save_features(self, identity_a, identity_b, features_dict):
        pair_hash = self._generate_pair_hash(identity_a, identity_b)
        features_json = json.dumps(features_dict)
        
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute(
            "INSERT OR REPLACE INTO features (pair_hash, features_json) VALUES (?, ?)",
            (pair_hash, features_json)
        )
        conn.commit()
        conn.close()

    def load_features(self, identity_a, identity_b):
        pair_hash = self._generate_pair_hash(identity_a, identity_b)
        
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()
        cursor.execute("SELECT features_json FROM features WHERE pair_hash = ?", (pair_hash,))
        result = cursor.fetchone()
        conn.close()
        return json.loads(result[0]) if result else None