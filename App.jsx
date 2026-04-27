import React, { useState, useEffect } from 'react';
import GraphView from './components/GraphView';
import TimelineView from './components/TimelineView';

function App() {
  const [graphData, setGraphData] = useState({ nodes: [], edges: [] });
  const [timelineData, setTimelineData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeView, setActiveView] = useState('graph'); // 'graph', 'alerts', 'behavior'

  // investigate current target
  const target = "shadow_coder";

  useEffect(() => {
    const fetchAnalysis = async () => {
      try {
        setLoading(true);
        // Fetch graph and timeline data from the Go backend concurrently
        const [gRes, tRes] = await Promise.all([
          fetch(`/api/graph?target=${target}`),
          fetch(`/api/timeline?target=${target}`)
        ]);

        if (!gRes.ok || !tRes.ok) throw new Error("Backend connection failed");

        const gData = await gRes.json();
        const tData = await tRes.json();

        setGraphData(gData);
        setTimelineData(tData);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchAnalysis();
  }, [target]);

  return (
    <div className="h-screen w-screen bg-[#050505] text-gray-200 flex overflow-hidden font-mono selection:bg-[#00ffcc]/30">
      {/* Left Navigation: Intelligence Sidebar */}
      <aside className="w-20 flex flex-col items-center py-8 border-r border-white/5 bg-[#08080a] z-30">
        <div className="mb-12 text-2xl filter drop-shadow-[0_0_8px_#00ffcc]">🕯️</div>
        <nav className="flex flex-col gap-8 flex-grow">
          {['graph', 'timeline', 'behavior', 'alerts'].map((view) => (
            <button
              key={view}
              onClick={() => setActiveView(view)}
              className={`group relative p-3 rounded-xl transition-all ${
                activeView === view ? 'bg-[#00ffcc]/10 text-[#00ffcc]' : 'text-gray-600 hover:text-gray-300'
              }`}
            >
              <div className="text-xs font-bold uppercase tracking-tighter transform -rotate-90 group-hover:rotate-0 transition-transform">
                {view[0]}
              </div>
              {activeView === view && (
                <div className="absolute right-0 top-1/4 w-1 h-1/2 bg-[#00ffcc] shadow-[0_0_10px_#00ffcc]" />
              )}
            </button>
          ))}
        </nav>
        <div className="mt-auto opacity-30 hover:opacity-100 transition-opacity">
          <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
        </div>
      </aside>

      <header className="fixed top-0 left-20 right-0 z-20 bg-[#050505]/80 backdrop-blur-md p-4 border-b border-white/5 flex justify-between items-center">
        <div className="flex items-center gap-4">
          <h1 className="text-sm font-black tracking-[0.3em] uppercase text-white">
            NOCTURNE <span className="text-[#00ffcc] ml-2 font-normal opacity-50">v1.0.4_ELITE</span>
          </h1>
          <div className="h-4 w-px bg-white/10" />
          <span className="text-[10px] text-gray-500 uppercase tracking-widest">Target: <span className="text-gray-300">{target}</span></span>
        </div>
        {loading && (
          <div className="flex items-center gap-3">
            <div className="text-[10px] text-[#00ffcc] animate-pulse uppercase tracking-widest">Decrypting_Pattern...</div>
            <div className="w-16 h-1 bg-white/5 rounded-full overflow-hidden">
              <div className="h-full bg-[#00ffcc] animate-[loading_2s_infinite]" />
            </div>
          </div>
        )}
      </header>
      
      <main className="flex-grow relative flex flex-col pt-16 ml-0">
        {/* Main Graph Visualization Area */}
        <div className="flex-grow relative">
          {activeView === 'graph' && <GraphView data={graphData} />}
          
          {/* Real-time Terminal Log Overlay */}
          <div className="absolute bottom-4 left-4 right-4 h-32 bg-black/40 border border-white/5 backdrop-blur-md rounded-lg overflow-hidden flex flex-col">
            <div className="bg-white/5 px-3 py-1 text-[9px] uppercase tracking-widest text-gray-500 flex justify-between">
              <span>Intelligence_Feed.log</span>
              <span className="text-[#00ffcc]">Live_Stream</span>
            </div>
            <div className="flex-grow p-3 text-[10px] font-mono overflow-y-auto space-y-1">
              <div className="text-blue-400">[INFO] <span className="text-gray-500">{new Date().toISOString()}</span> System initialization complete.</div>
              <div className="text-[#00ffcc]">[LINK] Verified high-confidence bridge: GitHub {`->`} Twitter</div>
              <div className="text-yellow-400 animate-pulse">[WARN] Behavior anomaly detected in timezone UTC+5.</div>
            </div>
          </div>
        </div>

        {/* Interactive Bottom Scrubber */}
        <footer className="h-1/4 min-h-[250px] bg-[#08080a] border-t border-white/10 shadow-[0_-10px_30px_rgba(0,0,0,0.5)] z-10 overflow-y-auto">
          <TimelineView timeline={timelineData} />
        </footer>
      </main>
    </div>
  );
}

export default App;