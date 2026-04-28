import numpy as np
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score
import logging
import yaml
import os
from datetime import datetime

from ml.dataset_builder import DatasetBuilder
from ml.feature_store import FeatureStore
from ml.model_loader import ModelLoader

# Setup logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def train_model():
    config_path = os.path.join(os.path.dirname(__file__), 'config.yaml')
    with open(config_path, 'r') as f:
        config = yaml.safe_load(f)

    logger.info("Starting ML model training pipeline...")

    # 1. Build Dataset
    dataset_builder = DatasetBuilder()
    identity_pairs, labels = dataset_builder.build_synthetic_dataset(num_samples=1000)
    logger.info(f"Generated {len(identity_pairs)} synthetic samples.")

    # 2. Extract Features and store in FeatureStore
    feature_store = FeatureStore(config['feature_store_db'])
    X = []
    y = []
    for i, (id_a, id_b) in enumerate(identity_pairs):
        features_dict = dataset_builder.feature_extractor.extract_features(id_a, id_b)
        feature_store.save_features(id_a, id_b, features_dict) # Cache features
        X.append(list(features_dict.values()))
        y.append(labels[i])
    
    X = np.array(X)
    y = np.array(y)
    logger.info(f"Extracted features for {len(X)} samples.")

    # 3. Train Model
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)
    
    model = LogisticRegression(solver='liblinear', random_state=42)
    model.fit(X_train, y_train)
    logger.info("Logistic Regression model trained.")

    # 4. Evaluate Model
    y_pred = model.predict(X_test)
    metrics = {
        "accuracy": accuracy_score(y_test, y_pred),
        "precision": precision_score(y_test, y_pred),
        "recall": recall_score(y_test, y_pred),
        "f1_score": f1_score(y_test, y_pred),
        "training_date": datetime.now().isoformat(),
        "version": config['model_version']
    }
    logger.info(f"Model evaluation metrics: {metrics}")

    # 5. Save Model and Metadata
    ModelLoader.save_model(model, config['model_dir'], config['model_filename'], config['metadata_filename'], metrics)
    logger.info("Model training pipeline completed.")

if __name__ == "__main__":
    train_model()