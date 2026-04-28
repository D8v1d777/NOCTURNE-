import os
import json
import pickle
import logging

logger = logging.getLogger(__name__)

class ModelLoader:
    """Handles loading and saving of ML models and their metadata."""

    @staticmethod
    def load_model(model_dir, model_filename, metadata_filename):
        model_path = os.path.join(model_dir, model_filename)
        metadata_path = os.path.join(model_dir, metadata_filename)

        if not os.path.exists(model_path) or not os.path.exists(metadata_path):
            logger.warning(f"Model or metadata not found at {model_path} or {metadata_path}. Returning None.")
            return None, None

        try:
            with open(model_path, 'rb') as f:
                model = pickle.load(f)
            with open(metadata_path, 'r') as f:
                metadata = json.load(f)
            logger.info(f"Model '{metadata.get('version', 'N/A')}' loaded successfully from {model_path}")
            return model, metadata
        except Exception as e:
            logger.error(f"Error loading model from {model_path}: {e}")
            return None, None

    @staticmethod
    def save_model(model, model_dir, model_filename, metadata_filename, metadata):
        os.makedirs(model_dir, exist_ok=True)
        model_path = os.path.join(model_dir, model_filename)
        metadata_path = os.path.join(model_dir, metadata_filename)
        try:
            with open(model_path, 'wb') as f:
                pickle.dump(model, f)
            with open(metadata_path, 'w') as f:
                json.dump(metadata, f, indent=2)
            logger.info(f"Model '{metadata.get('version', 'N/A')}' saved successfully to {model_path}")
        except Exception as e:
            logger.error(f"Error saving model to {model_path}: {e}")