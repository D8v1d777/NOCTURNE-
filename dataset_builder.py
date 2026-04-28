import numpy as np
import random
from ml.feature_extractor import FeatureExtractor

class DatasetBuilder:
    """Generates synthetic datasets for training the ML identity scorer."""

    def __init__(self):
        self.feature_extractor = FeatureExtractor()

    def _generate_identity(self, base_username, base_email, base_bio, base_links, base_avatar_hash, platform_suffix=""):
        """Generates a slightly varied identity based on a base."""
        return {
            "id": f"id_{random.randint(1000, 9999)}",
            "username": base_username + (random.choice(["", "_dev", "_osint"]) if random.random() < 0.3 else ""),
            "email": base_email if random.random() < 0.9 else f"alt_{base_email}",
            "bio": base_bio + (random.choice(["", " Loves Go.", " Security researcher."]) if random.random() < 0.4 else ""),
            "avatar_hash": base_avatar_hash if random.random() < 0.9 else f"a{random.randint(1000, 9999)}",
            "platform": f"platform_{platform_suffix or random.choice(['github', 'twitter', 'reddit'])}",
            "links": base_links if random.random() < 0.8 else base_links + [f"https://newlink.com/{random.randint(1,100)}"]
        }

    def build_synthetic_dataset(self, num_samples=100):
        """
        Builds a synthetic dataset of identity pairs and their true match labels.
        Returns: (list of (identity_a, identity_b), list of labels)
        """
        identity_pairs = []
        labels = []

        base_identities = [
            ("shadow_coder", "shadow@example.com", "OSINT researcher and Go enthusiast", ["https://shadow.io"], "a1b2c3d4"),
            ("john_doe", "john.doe@email.com", "Software engineer, Python lover", ["https://johndoe.dev"], "b2c3d4e5"),
            ("jane_smith", "jane.smith@mail.com", "Data scientist, ML practitioner", ["https://janesmith.ai"], "c3d4e5f6"),
        ]

        for _ in range(num_samples // 2): # Generate positive samples
            base_user, base_email, base_bio, base_links, base_avatar = random.choice(base_identities)
            
            id_a = self._generate_identity(base_user, base_email, base_bio, base_links, base_avatar, "github")
            id_b = self._generate_identity(base_user, base_email, base_bio, base_links, base_avatar, "twitter")
            
            identity_pairs.append((id_a, id_b))
            labels.append(1)

        for _ in range(num_samples // 2): # Generate negative samples
            # Pick two different base identities
            (user1, email1, bio1, links1, avatar1) = random.choice(base_identities)
            (user2, email2, bio2, links2, avatar2) = random.choice([bi for bi in base_identities if bi[0] != user1])

            id_a = self._generate_identity(user1, email1, bio1, links1, avatar1, "github")
            id_b = self._generate_identity(user2, email2, bio2, links2, avatar2, "reddit")
            
            identity_pairs.append((id_a, id_b))
            labels.append(0)

        return identity_pairs, labels

    def build_dataset_from_feature_store(self, feature_store_db_path):
        """
        (Placeholder for future implementation)
        Builds a dataset from features stored in the SQLite feature store.
        This would require a mechanism to store the true labels alongside features.
        """
        raise NotImplementedError("Building dataset from feature store is not yet implemented.")

if __name__ == "__main__":
    builder = DatasetBuilder()
    pairs, labels = builder.build_synthetic_dataset(num_samples=10)
    print(f"Generated {len(pairs)} synthetic samples.")
    for i, (id_a, id_b) in enumerate(pairs):
        print(f"\nPair {i+1} (Label: {labels[i]}):")
        print(f"  ID A: {id_a['username']} ({id_a['platform']})")
        print(f"  ID B: {id_b['username']} ({id_b['platform']})")
        features = builder.feature_extractor.extract_features(id_a, id_b)
        print(f"  Features: {features}")