import json
import requests
import os
import sys

class QueryBuilder:
    """Dynamically constructs and validates search queries based on source capabilities."""
    
    def __init__(self, sources_path="sources.json", registry_path="sources_registry.json"):
        self.base_dir = os.path.dirname(os.path.abspath(__file__))
        self.sources_path = os.path.join(self.base_dir, sources_path)
        self.registry_path = os.path.join(self.base_dir, registry_path)
        
        self.sources = self._load_json(self.sources_path, [])
        self.registry = self._load_json(self.registry_path, {})

    def _load_json(self, path, default):
        if not os.path.exists(path):
            return default
        try:
            with open(path, 'r', encoding='utf-8') as f:
                return json.load(f)
        except Exception as e:
            sys.stderr.write(f"Error loading {path}: {e}\n")
            return default

    def build_queries(self, query_data):
        """
        Selects appropriate sources from registry and builds validated query URLs.
        Input format: {"type": "person", "name": "John Smith"}
        """
        results = []
        name = query_data.get("name")
        email = query_data.get("email")

        # Determine the required capability based on the input field provided
        required_cap = None
        target_value = None
        if name:
            required_cap = "supports_name"
            target_value = name
        elif email:
            required_cap = "supports_email"
            target_value = email

        if not required_cap:
            return results

        # Create a lookup map for source templates from sources.json
        templates = {s['id']: s.get('url_template') for s in self.sources}

        # Filter the registry for sources that support the required capability
        for source_id, entry in self.registry.items():
            capabilities = entry.get("capabilities", {})
            
            # Shadow Registry Logic: Keep 'broken' or 'blocked' as fallback candidates
            is_shadow = entry.get("status") in ["broken", "blocked", "login_required"]
            
            if capabilities.get(required_cap) and (entry.get("status") == "active" or (include_shadow and is_shadow)):
                url_template = templates.get(source_id)
                if not url_template:
                    continue

                # Construct URL using the template (no hardcoding)
                query_url = url_template.format(query=target_value)
                
                # Validate by sending a lightweight test request
                is_valid = self._validate_request(source_id, query_url)
                
                results.append({
                    "source": source_id,
                    "query_url": query_url,
                    "valid": is_valid
                })

        return results

    def _validate_request(self, source_id, url):
        """Verifies if the query URL is operational using the source's defined method."""
        source_cfg = next((s for s in self.sources if s['id'] == source_id), {})
        method = source_cfg.get('method', 'GET')
        
        try:
            # Use stream=True to avoid downloading the full body during validation
            resp = requests.request(
                method=method,
                url=url,
                timeout=5,
                headers={"User-Agent": "Nocturne-Validator/1.0"},
                stream=True
            )
            return resp.status_code < 400
        except Exception:
            return False

if __name__ == "__main__":
    # Example usage: python query_builder.py '{"type": "person", "name": "John Smith"}'
    sample_input = {"type": "person", "name": "John Smith"}
    
    if len(sys.argv) > 1:
        try:
            sample_input = json.loads(sys.argv[1])
        except: pass

    qb = QueryBuilder()
    print(json.dumps(qb.build_queries(sample_input), indent=2))