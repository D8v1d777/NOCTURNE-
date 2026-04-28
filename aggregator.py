import json
import time
from distributed.queue import NocturneQueue
# from ml.scorer import MLScorer # Integrated with previous intelligence tasks

class ResultAggregator:
    """Collects and processes results from the distributed workers."""
    
    def __init__(self):
        self.queue = NocturneQueue()
        # self.scorer = MLScorer() 

    def start(self):
        print("[*] Result Aggregator active. Listening for worker outputs...")
        while True:
            result = self.queue.pop_result()
            if result:
                self.process_result(result)

    def process_result(self, result):
        source_id = result.get("source_id")
        target = result.get("target")
        
        if result["status"] == "success":
            identities = result.get("identities", [])
            print(f"[AGGR] Received {len(identities)} identities from {source_id} for {target}")
            # Here we would call self.scorer.score_identity_match(...) 
            # and push to the final Graph API / database
        else:
            print(f"[AGGR] Job failed for {source_id}: {result.get('error')}")

if __name__ == "__main__":
    ResultAggregator().start()