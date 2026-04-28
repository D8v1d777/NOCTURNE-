import json
import re

class SourceClassifier:
    """Determines content type, source category, and technical requirements."""
    
    @staticmethod
    def classify(source_id, content):
        classification = {
            "source_id": source_id,
            "content_type": "html",
            "source_type": "web",
            "contains_fields": [],
            "pagination": False,
            "js_required": False,
            "confidence": 0.5
        }
        
        if not content:
            return classification

        # 1. Content Type Detection
        try:
            json.loads(content)
            classification["content_type"] = "json"
            classification["source_type"] = "api"
            classification["confidence"] += 0.2
        except (ValueError, TypeError):
            pass

        # 2. Field Hint Detection
        field_hints = {
            "username": [r"user", r"handle", r"alias", r"login"],
            "email": [r"email", r"contact", r"mailto"],
            "company": [r"company", r"corp", r"inc", r"officer"],
            "links": [r"http", r"www", r"url"]
        }
        
        content_lower = content[:50000].lower() # Sample start for efficiency
        for field, patterns in field_hints.items():
            if any(re.search(p, content_lower) for p in patterns):
                classification["contains_fields"].append(field)

        # 3. Pagination Detection
        pagination_patterns = [r"page=", r"offset=", r"next", r"pagination", r"total_pages"]
        if any(re.search(p, content_lower) for p in pagination_patterns):
            classification["pagination"] = True

        # 4. JS Requirement Detection (Heuristic)
        if classification["content_type"] == "html":
            scripts = len(re.findall(r'<script', content_lower))
            visible_text_len = len(re.sub(r'<[^>]*>', '', content_lower))
            
            # High script-to-text ratio or presence of SPA markers
            if (scripts > 5 and visible_text_len < 2000) or "noscript" in content_lower:
                classification["js_required"] = True
                classification["confidence"] += 0.1

        # Normalize confidence
        classification["confidence"] = round(min(1.0, classification["confidence"]), 2)
        
        return classification