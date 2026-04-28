import re

class Scorer:
    KEYWORDS = {
        "username": ["handle", "user", "alias", "profile", "who"],
        "email": ["email", "contact", "mailto", "address"],
        "name": ["name", "officer", "owner", "full name"],
        "bio": ["bio", "about", "description", "summary"],
        "location": ["location", "city", "country", "address"]
    }

    @staticmethod
    def check_proximity(field, context):
        if not context:
            return False
        keywords = Scorer.KEYWORDS.get(field, [])
        return any(k in context.lower() for k in keywords)

    @staticmethod
    def calculate_field_confidence(field, has_pattern, has_proximity, structural_boost):
        score = 0.4 if has_pattern else 0.0
        if has_proximity:
            score += 0.35
        score += structural_boost
        
        return round(min(0.98, score), 2)

    @staticmethod
    def aggregate_identity_confidence(fields):
        if not fields:
            return 0.0
        avg = sum(f["confidence"] for f in fields) / len(fields)
        return round(min(1.0, avg), 2)