import subprocess
import time
import os
import sys
import json

# --- Configuration ---
ROOT_DIR = os.path.dirname(os.path.abspath(__file__))

# Go API Server
GO_API_PATH = os.path.join(ROOT_DIR, "modules_go", "concurrent_scanner")
GO_API_COMMAND = ["go", "run", "main.go", "serve"]

# Python Distributed Components
PYTHON_EXECUTABLE = sys.executable # Use the current Python interpreter
WORKER_PATH = os.path.join(ROOT_DIR, "distributed", "worker.py")
AGGREGATOR_PATH = os.path.join(ROOT_DIR, "distributed", "aggregator.py")
CONTROLLER_PATH = os.path.join(ROOT_DIR, "distributed", "controller.py")

# ML Training (optional, run once)
ML_TRAIN_PATH = os.path.join(ROOT_DIR, "ml", "train.py")

# Default target for controller and scheduler
DEFAULT_TARGET_NAME = "johnsmith"
DEFAULT_SCHEDULER_QUERY = json.dumps({"name": DEFAULT_TARGET_NAME})

# List to keep track of all started processes
running_processes = []

def check_exists(path):
    """Safety check for component existence."""
    if not os.path.exists(path):
        print(f"[ERROR] Missing critical component: {path}")
        return False
    return True

def start_component(name, cmd, cwd=None, env=None):
    """Starts a component as a background process."""
    print(f"[*] Starting {name}...")
    process = subprocess.Popen(
        cmd,
        cwd=cwd,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        bufsize=1 # Line-buffered output
    )
    running_processes.append(process)
    print(f"[*] {name} started with PID: {process.pid}")
    # Give it a moment to start up and print initial logs
    time.sleep(1)
    return process

def stream_logs(process, prefix):
    """Streams stdout and stderr of a process to the console."""
    def _stream(pipe, stream_type):
        for line in pipe:
            print(f"[{prefix}-{stream_type}] {line.strip()}")
    
    # Start streaming in separate threads to avoid blocking
    import threading
    threading.Thread(target=_stream, args=(process.stdout, "OUT"), daemon=True).start()
    threading.Thread(target=_stream, args=(process.stderr, "ERR"), daemon=True).start()

def stop_all_components():
    """Terminates all started background processes."""
    print("\n[*] Shutting down all NOCTURNE components...")
    for p in reversed(running_processes):
        if p.poll() is None: # Only terminate if still running
            print(f"[*] Terminating PID: {p.pid}")
            p.terminate()
            try:
                p.wait(timeout=5) # Wait for process to exit gracefully
            except subprocess.TimeoutExpired:
                print(f"[!] PID {p.pid} did not terminate gracefully, killing...")
                p.kill()
    print("[*] All components shut down.")

def main():
    target_name = DEFAULT_TARGET_NAME
    if len(sys.argv) > 1:
        target_name = sys.argv[1]
    
    print("--- NOCTURNE Automated Startup ---")
    print(f"Targeting: {target_name}")
    print("----------------------------------")

    try:
        # 1. Start Go API Server
        if check_exists(os.path.join(GO_API_PATH, "main.go")):
            go_api_proc = start_component("Go API Server", GO_API_COMMAND, cwd=GO_API_PATH)
            stream_logs(go_api_proc, "GO_API")
            time.sleep(3) # Give Go server time to fully initialize

        # 2. Train ML Model (if not already trained)
        # This is a one-off task, not a continuous service
        print("[*] Checking ML model status. Training if necessary...")
        if check_exists(ML_TRAIN_PATH):
            subprocess.run([PYTHON_EXECUTABLE, ML_TRAIN_PATH], cwd=ROOT_DIR)
            print("[*] ML model ready.")

        # 3. Start Aggregator
        if check_exists(AGGREGATOR_PATH):
            aggregator_proc = start_component("Result Aggregator", [PYTHON_EXECUTABLE, AGGREGATOR_PATH], cwd=ROOT_DIR)
            stream_logs(aggregator_proc, "AGGR")

        # 4. Start Worker (can run multiple instances by calling this multiple times)
        if check_exists(WORKER_PATH):
            worker_proc = start_component("Worker (US/html)", [PYTHON_EXECUTABLE, WORKER_PATH], cwd=ROOT_DIR, env={**os.environ, "REGION": "US", "CAPABILITY": "html"})
            stream_logs(worker_proc, "WORKER")

        # 5. Dispatch Scan Jobs via Controller
        print(f"\n[*] Dispatching scan for target '{target_name}' via Controller...")
        if check_exists(CONTROLLER_PATH):
            controller_proc = start_component("Controller", [PYTHON_EXECUTABLE, CONTROLLER_PATH, target_name], cwd=ROOT_DIR)
            stream_logs(controller_proc, "CTRL")
        
        # Keep the main script alive indefinitely, or until Ctrl+C
        print("\n[*] NOCTURNE is running. Press Ctrl+C to stop all components.")
        while True:
            time.sleep(1)

    except KeyboardInterrupt:
        print("\n[*] Ctrl+C detected. Initiating graceful shutdown...")
    except Exception as e:
        print(f"[CRITICAL] An error occurred: {e}")
    finally:
        stop_all_components()

if __name__ == "__main__":
    # Add ROOT_DIR to sys.path so Python can find 'core', 'distributed', 'ml' modules
    sys.path.insert(0, ROOT_DIR)
    main()