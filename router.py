class JobRouter:
    """Decides which worker type and region should handle a specific job."""

    @staticmethod
    def get_routing_metadata(source_id, classifier_output, profile_manager):
        """
        Returns metadata for queue routing.
        """
        # 1. Determine Capability Requirement
        # If classifier detected high JS dependency, route to JS workers
        required_cap = "js" if classifier_output.get("js_required") else "html"
        
        # 2. Determine Region Preference
        preferred_region = profile_manager.get_source_region_preference(source_id)

        # 3. Calculate Node Priority (Heuristic)
        # Higher priority for workers in the preferred region
        
        return {
            "required_capability": required_cap,
            "preferred_region": preferred_region,
            "routing_key": f"nocturne:jobs:{preferred_region}:{required_cap}"
        }

    @staticmethod
    def get_fallback_keys(region, capability):
        """If preferred region/cap is empty, provide fallback keys."""
        fallbacks = [
            f"nocturne:jobs:ANY:{capability}",
            f"nocturne:jobs:ANY:html" # Final fallback
        ]
        if region != "ANY":
            fallbacks.insert(0, f"nocturne:jobs:{region}:html")
            
        return fallbacks