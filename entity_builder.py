import hashlib
from core.scoring import Scorer

class EntityBuilder:
    """Merges detected fields into unique Identity objects."""

    @staticmethod
    def build(source_id, detected_fields):
        if not detected_fields:
            return []

        # Group by field types for merging (e.g., take highest confidence email)
        best_fields = {}
        links = []

        for df in detected_fields:
            field = df["field"]
            if field == "links":
                links.append(df["value"])
                continue
                
            if field not in best_fields or df["confidence"] > best_fields[field]["confidence"]:
                best_fields[field] = df

        # Generate ID based on primary unique signals
        seed = f"{source_id}:{best_fields.get('username', {'value': ''})['value']}:{best_fields.get('email', {'value': ''})['value']}"
        identity_id = hashlib.sha256(seed.encode()).hexdigest()[:16]

        identity = {
            "id": f"idnt_{identity_id}",
            "username": best_fields.get("username", {}).get("value"),
            "email": best_fields.get("email", {}).get("value"),
            "name": best_fields.get("name", {}).get("value"),
            "bio": best_fields.get("bio", {}).get("value"),
            "location": best_fields.get("location", {}).get("value"),
            "links": list(set(links)),
            "source": source_id,
            "confidence": Scorer.aggregate_identity_confidence(list(best_fields.values()))
        }

        return [identity]