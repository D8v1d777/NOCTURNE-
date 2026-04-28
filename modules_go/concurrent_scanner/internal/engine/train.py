import os

def train():
    print("[ML] Training model (stub)...")
    os.makedirs("ml/models", exist_ok=True)

    # Fake model file
    with open("ml/models/model.bin", "w") as f:
        f.write("dummy-model")

    print("[ML] Model saved.")

if __name__ == "__main__":
    train()