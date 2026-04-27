import React, { useEffect, useRef, useState, useMemo } from 'react';
import cytoscape from 'cytoscape';

const PLATFORM_COLORS = {
  github: '#fafbfc',
  twitter: '#1DA1F2',
  reddit: '#FF4500',
  instagram: '#E1306C',
  mastodon: '#2b90d9',
  bluesky: '#0560ff',
  default: '#8884d8'
};

const GraphView = ({ data }) => {
  const containerRef = useRef(null);
  const cyRef = useRef(null);
  const [selectedNode, setSelectedNode] = useState(null);
  const [threshold, setThreshold] = useState(0.65);
  const [filterPlatform, setFilterPlatform] = useState('all');

  // Filter data based on UI controls
  const filteredData = useMemo(() => {
    const nodes = data.nodes.filter(n => filterPlatform === 'all' || n.platform.toLowerCase() === filterPlatform.toLowerCase());
    const nodeIds = new Set(nodes.map(n => n.id));
    
    const edges = data.edges.filter(e => 
      e.weight >= threshold && 
      nodeIds.has(e.source) && 
      nodeIds.has(e.target)
    );

    return { nodes, edges };
  }, [data, threshold, filterPlatform]);

  useEffect(() => {
    if (!containerRef.current) return;

    // Initialize Cytoscape
    const cy = cytoscape({
      container: containerRef.current,
      elements: [
        ...filteredData.nodes.map(n => ({
          data: { ...n, color: PLATFORM_COLORS[n.platform.toLowerCase()] || PLATFORM_COLORS.default }
        })),
        ...filteredData.edges.map(e => ({
          data: { ...e }
        }))
      ],
      style: [
        {
          selector: 'node',
          style: {
            'background-color': '#000',
            'label': 'data(label)',
            'color': '#888',
            'font-size': '10px',
            'font-family': 'monospace',
            'text-valign': 'bottom',
            'text-margin-y': 8,
            'width': 'mapData(confidence, 0, 1, 15, 45)',
            'height': 'mapData(confidence, 0, 1, 15, 45)',
            'border-width': 1.5,
            'border-color': 'data(color)',
            'transition-property': 'background-color, line-color, border-width, opacity, text-opacity',
            'transition-duration': '0.5s'
          }
        },
        {
          selector: 'edge',
          style: {
            'width': 'mapData(weight, 0.6, 1, 0.5, 3)',
            'line-color': '#222',
            'curve-style': 'bezier',
            'opacity': 0.4,
            'overlay-opacity': 0
          }
        },
        {
          selector: 'node[confidence > 0.9]',
          style: {
            'border-width': 3,
            'shadow-blur': 10,
            'shadow-color': 'data(color)',
            'shadow-opacity': 0.8
          }
        },
        {
          selector: 'node:selected',
          style: {
            'background-color': 'data(color)',
            'border-color': '#fff',
            'border-width': 2,
            'text-opacity': 1,
            'color': '#fff'
          }
        },
        {
          selector: '.faded',
          style: {
            'opacity': 0.1,
            'text-opacity': 0
          }
        },
        {
          selector: '.highlighted',
          style: {
            'opacity': 1,
            'text-opacity': 1,
            'line-color': '#00ffcc'
          }
        }
      ],
      layout: {
        name: 'cose',
        animate: true,
        randomize: true,
        componentSpacing: 100,
        nodeRepulsion: 400000,
        edgeElasticity: 100
      }
    });

    // High-End Interaction: Active Cluster Highlighting
    cy.on('tap', 'node', (evt) => {
      const node = evt.target;
      setSelectedNode(node.data());
      
      // Fade unrelated nodes
      cy.elements().addClass('faded').removeClass('highlighted');
      node.closedNeighborhood().addClass('highlighted').removeClass('faded');
    });

    cy.on('tap', (evt) => {
      if (evt.target === cy) {
        setSelectedNode(null);
        cy.elements().removeClass('faded highlighted');
      }
    });

    // Pulsing Animation for high-confidence nodes
    cy.nodes('[confidence > 0.9]').forEach(n => {
        n.animation({ style: { 'border-width': 6, 'shadow-blur': 20 }, duration: 1500 }).play()
         .promise().then(() => n.animation({ style: { 'border-width': 3, 'shadow-blur': 10 }, duration: 1500 }).play());
    });

    cyRef.current = cy;

    return () => cy.destroy();
  }, [filteredData]);

  return (
    <div className="flex h-screen bg-[#0a0a0c] text-gray-100 overflow-hidden font-sans">
      {/* Main Graph Canvas */}
      <div className="relative flex-grow cursor-crosshair">
        <div ref={containerRef} className="w-full h-full" />
        
        {/* Floating Controls */}
        <div className="absolute top-20 left-4 p-4 bg-[#08080a]/90 border border-white/5 rounded-lg backdrop-blur-md z-10 w-64 shadow-2xl">
          <h3 className="text-xs font-bold uppercase tracking-widest text-[#00ffcc] mb-4">Correlation Filters</h3>
          
          <div className="mb-4">
            <label className="block text-[10px] uppercase text-gray-500 mb-1">Min Confidence: {threshold.toFixed(2)}</label>
            <input 
              type="range" min="0.4" max="1.0" step="0.05" 
              value={threshold} 
              onChange={(e) => setThreshold(parseFloat(e.target.value))}
              className="w-full h-1 bg-gray-700 rounded-lg appearance-none cursor-pointer accent-[#00ffcc]"
            />
          </div>

          <div>
            <label className="block text-[10px] uppercase text-gray-500 mb-1">Platform Filter</label>
            <select 
              className="w-full bg-[#1c1c21] border border-white/10 rounded px-2 py-1 text-sm outline-none"
              onChange={(e) => setFilterPlatform(e.target.value)}
            >
              <option value="all">All Platforms</option>
              <option value="github">GitHub</option>
              <option value="twitter">Twitter</option>
              <option value="reddit">Reddit</option>
            </select>
          </div>
        </div>
      </div>

      {/* Side Detail Panel */}
      <div className={`w-96 bg-[#0f0f12] border-l border-white/10 p-6 transition-all transform ${selectedNode ? 'translate-x-0' : 'translate-x-full'}`}>
        {selectedNode ? (
          <div className="space-y-6">
            <div className="flex items-center space-x-4">
              <div 
                className="w-12 h-12 rounded-full border-2 border-[#00ffcc] flex items-center justify-center text-xl font-bold"
                style={{ backgroundColor: selectedNode.color }}
              >
                {selectedNode.platform[0]}
              </div>
              <div>
                <h2 className="text-xl font-bold">{selectedNode.label}</h2>
                <p className="text-xs text-gray-500 uppercase tracking-tighter">{selectedNode.platform}</p>
              </div>
            </div>

            <div className="p-3 bg-[#16161a] rounded-md border border-white/5">
              <p className="text-[10px] uppercase text-gray-500 mb-1">Confidence Score</p>
              <div className="flex items-center space-x-2">
                <div className="flex-grow h-1.5 bg-gray-800 rounded-full overflow-hidden">
                  <div className="h-full bg-[#00ffcc]" style={{ width: `${selectedNode.confidence * 100}%` }} />
                </div>
                <span className="text-xs font-mono">{(selectedNode.confidence * 100).toFixed(0)}%</span>
              </div>
            </div>

            <div>
              <h4 className="text-xs font-bold uppercase text-gray-400 mb-2">Profile Metadata</h4>
              <div className="text-sm space-y-2 text-gray-300">
                <div className="flex justify-between">
                  <span className="text-gray-600 italic">ID:</span>
                  <span className="font-mono text-xs">{selectedNode.id}</span>
                </div>
                <p className="text-sm italic text-gray-400">"{selectedNode.bio || 'No bio provided'}"</p>
              </div>
            </div>

            <button 
              onClick={() => setSelectedNode(null)}
              className="w-full py-2 border border-white/10 hover:bg-white/5 rounded text-xs transition-colors"
            >
              Close Analysis
            </button>
          </div>
        ) : (
          <div className="h-full flex flex-col items-center justify-center text-center opacity-20">
            <div className="mb-4 text-4xl">🕸️</div>
            <p className="text-sm uppercase tracking-widest">Select an identity node<br/>to begin extraction</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default GraphView;