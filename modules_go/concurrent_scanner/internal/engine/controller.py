import sys
import time

if __name__ == "__main__":
    target = sys.argv[1] if len(sys.argv) > 1 else "unknown"

    print(f"[CONTROLLER] Dispatching scan for: {target}")

    # simulate dispatch
    time.sleep(2)

    print(f"[CONTROLLER] Tasks sent to workers")