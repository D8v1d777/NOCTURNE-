import React, { useState } from 'react';

const EVENT_STYLES = {
  account_created: 'bg-emerald-500 border-emerald-400',
  post_activity: 'bg-blue-500 border-blue-400',
  profile_update: 'bg-amber-500 border-amber-400',
  link_found: 'bg-purple-500 border-purple-400'
};

const TimelineView = ({ timeline }) => {
  const [selectedEvent, setSelectedEvent] = useState(null);

  if (!timeline || !timeline.events.length) return null;

  // Calculate positions based on timestamps
  const startTime = new Date(timeline.events[0].timestamp).getTime();
  const endTime = new Date(timeline.events[timeline.events.length - 1].timestamp).getTime();
  const duration = endTime - startTime || 1;

  return (
    <div className="bg-transparent p-6 w-full">
      {/* Strategic Insight Panel */}
      {timeline.strategic_insights && timeline.strategic_insights.length > 0 && (
        <div className="mb-8 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {timeline.strategic_insights.map((insight, i) => {
            const colorMap = {
              Stability: 'text-emerald-400 border-emerald-500/20 bg-emerald-500/5',
              Risk: 'text-red-400 border-red-500/20 bg-red-500/5',
              Growth: 'text-blue-400 border-blue-500/20 bg-blue-500/5',
              Activity: 'text-purple-400 border-purple-500/20 bg-purple-500/5',
            };
            const colors = colorMap[insight.category] || 'text-gray-400 border-white/5 bg-white/5';
            
            return (
              <div key={i} className={`p-4 border rounded-lg backdrop-blur-sm transition-all hover:scale-[1.02] ${colors}`}>
                <div className="flex justify-between items-center mb-2">
                  <span className="text-[10px] font-black uppercase tracking-[0.2em]">{insight.category}</span>
                  <div className="flex items-center gap-1">
                    <div className="w-1.5 h-1.5 rounded-full bg-current animate-pulse" />
                    <span className="text-[9px] font-mono opacity-60">{(insight.confidence * 100).toFixed(0)}%</span>
                  </div>
                </div>
                <p className="text-[11px] leading-relaxed font-bold tracking-tight text-white/90">
                  {insight.message}
                </p>
                
                {/* Minimal Confidence Vector */}
                <div className="mt-3 h-0.5 w-full bg-white/5 rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-current opacity-40 transition-all duration-1000" 
                      style={{ width: `${insight.confidence * 100}%` }}
                    />
                </div>
              </div>
            );
          })}
        </div>
      )}

      <div className="flex justify-between items-center mb-8">
        <div>
          <h3 className="text-xs font-black uppercase tracking-[0.4em] text-[#00ffcc] opacity-80">Chronological_Analysis</h3>
          <p className="text-xs text-gray-500">Chronological identity progression</p>
        </div>
        <div className="flex gap-2">
          {Object.keys(EVENT_STYLES).map(type => (
            <div key={type} className="flex items-center gap-1">
              <div className={`w-2 h-2 rounded-full ${EVENT_STYLES[type].split(' ')[0]}`} />
              <span className="text-[10px] text-gray-400 capitalize">{type.replace('_', ' ')}</span>
            </div>
          ))}
        </div>
      </div>

      {/* AI Summary Section */}
      {timeline.summary ? (
        <div className="mb-8 p-4 bg-[#16161a] border-l-4 border-[#00ffcc] rounded-r-md shadow-inner relative overflow-hidden">
          {/* Uncertainty Overlay */}
          {timeline.summary.risk_flags.some(f => f.includes('UNCERTAINTY')) && (
            <div className="absolute top-0 right-0 p-1 bg-red-500 text-[8px] font-black text-white uppercase tracking-tighter animate-pulse">
              Low_Confidence_Warning
            </div>
          )}
          
          <div className="flex justify-between items-start mb-4">
            <h4 className="text-[10px] uppercase text-gray-500 font-bold tracking-widest">Identity Intelligence Summary</h4>
            <div className="w-48">
              <div className="flex justify-between text-[8px] uppercase text-gray-600 mb-1">
                <span>Confidence_Vector</span>
                <span>{(timeline.summary.score * 100).toFixed(0)}%</span>
              </div>
              <div className="h-1 w-full bg-white/5 rounded-full">
                <div 
                  className={`h-full transition-all duration-1000 ${timeline.summary.score > 0.8 ? 'bg-[#00ffcc]' : 'bg-yellow-500'}`} 
                  style={{ width: `${timeline.summary.score * 100}%` }}
                />
              </div>
            </div>
          </div>

          <p className="text-sm text-gray-200 leading-relaxed mb-4">
            {timeline.summary.summary}
          </p>

          <div className="flex flex-wrap gap-8">
            <div className="space-y-1">
              <p className="text-[9px] uppercase text-gray-500 font-bold">Key Insights</p>
              {timeline.summary.key_points.map((pt, i) => (
                <div key={i} className="flex items-center gap-2 text-[11px] text-emerald-400">
                  <span className="opacity-50">•</span> {pt}
                </div>
              ))}
            </div>
            {/* Evidence Panel */}
            <div className="space-y-1 border-l border-white/5 pl-8 flex-grow">
                <p className="text-[9px] uppercase text-gray-500 font-bold">Hard Evidence</p>
                {timeline.summary.evidence.map((ev, i) => (
                    <div key={i} className="flex items-center gap-2 text-[11px] text-blue-400">
                        <span className="text-[8px] opacity-50">▶</span> {ev}
                    </div>
                ))}
            </div>
            {timeline.summary.risk_flags.length > 0 && (
              <div className="space-y-1 border-l border-white/5 pl-8">
                <p className="text-[9px] uppercase text-gray-500 font-bold">Risk Indicators</p>
                {timeline.summary.risk_flags.map((flag, i) => (
                  <div key={i} className="flex items-center gap-2 text-[11px] text-red-400">
                    <span>⚠️</span> {flag}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="mb-8 p-8 border border-white/5 border-dashed text-center text-gray-600 text-xs uppercase tracking-widest">Awaiting Cluster Synthesis...</div>
      )}

      {/* Predictive Intelligence Section */}
      {timeline.predictions && timeline.predictions.length > 0 && (
        <div className="mb-8 grid grid-cols-1 md:grid-cols-2 gap-4">
          {timeline.predictions.map((pred, i) => (
            <div key={i} className="p-3 bg-blue-500/5 border border-blue-500/10 rounded-md relative group hover:border-blue-500/30 transition-colors">
              <div className="flex justify-between items-center mb-1">
                <span className="text-[10px] uppercase text-blue-400/80 font-black tracking-widest">Future Signal (Estimated)</span>
                <span className="text-[10px] text-gray-500 font-mono">{(pred.probability * 100).toFixed(0)}% Confidence</span>
              </div>
              <p className="text-sm font-bold text-gray-200 capitalize tracking-tight">{pred.type.replace('_', ' ')}</p>
              <p className="text-[11px] text-gray-500 mt-1 italic">"{pred.reason}"</p>
              {pred.estimated_time && (
                <div className="mt-2 pt-2 border-t border-white/5">
                  <span className="text-[10px] text-blue-300/60 font-mono">Projected Window: {new Date(pred.estimated_time).toLocaleDateString()}</span>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Identity Evolution Log */}
      {timeline.history && timeline.history.length > 0 && (
        <div className="mb-8 p-4 bg-[#1a1a20]/50 border border-white/5 rounded-md">
          <h4 className="text-[10px] uppercase text-gray-500 font-bold mb-4 tracking-widest">Identity_Evolution_Log</h4>
          <div className="space-y-4">
            {timeline.history.map((change, i) => (
              <div key={i} className="flex gap-4 items-start text-xs">
                <div className="w-24 text-gray-500 font-mono text-[10px]">
                  {new Date(change.timestamp).toLocaleDateString()}
                </div>
                <div className="flex-grow">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="px-1.5 py-0.5 rounded bg-blue-500/20 text-blue-400 border border-blue-500/30 text-[9px] uppercase font-bold">
                      {change.type.replace('_', ' ')}
                    </span>
                    <span className="text-gray-400 font-bold">{change.platform}</span>
                  </div>
                  <p className="text-gray-300">
                    {change.type === 'username_change' ? (
                      <>Rebranded from <span className="text-red-400/80">'{change.old}'</span> to <span className="text-[#00ffcc]">'{change.new}'</span></>
                    ) : change.type === 'bio_update' ? (
                      <>Profile bio updated</>
                    ) : (
                      <>{change.type.replace('_', ' ')}: {change.new}</>
                    )}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Behavioral Profile Summary */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
        <div className="p-3 bg-[#1a1a20] border border-white/5 rounded-md">
          <p className="text-[10px] uppercase text-gray-500 mb-1">Behavior Pattern</p>
          <p className="text-sm font-bold text-[#00ffcc] capitalize">{timeline.profile?.activity_pattern?.replace('_', ' ') || 'Calculating...'}</p>
        </div>
        <div className="p-3 bg-[#1a1a20] border border-white/5 rounded-md">
          <p className="text-[10px] uppercase text-gray-500 mb-1">Geo Inference</p>
          <div className="flex flex-col">
            <p className="text-sm font-bold text-gray-100">
              {timeline.profile?.geo?.probable_region}
            </p>
            <span className="text-[9px] text-[#00ffcc] font-mono">{timeline.profile?.geo?.probable_timezone} ({(timeline.profile?.geo?.confidence * 100).toFixed(0)}% conf)</span>
          </div>
        </div>
        <div className="p-3 bg-[#1a1a20] border border-white/5 rounded-md">
          <p className="text-[10px] uppercase text-gray-500 mb-1">Platform Spread</p>
          <div className="flex gap-1">
            {Object.keys(timeline.profile?.platform_usage || {}).map(p => (
              <span key={p} className="text-[9px] bg-white/5 px-1.5 rounded border border-white/5" title={p}>{p[0]}</span>
            ))}
          </div>
        </div>
      </div>

      {/* Insight Badges */}
      <div className="flex flex-wrap gap-2 mb-10">
        {timeline.insights.map((insight, i) => {
          const isAnomaly = timeline.profile?.anomalies?.includes(insight);
          return (
            <div key={i} className={`px-3 py-1 border rounded-full text-[11px] italic transition-colors ${
              isAnomaly ? 'bg-red-500/10 border-red-500/30 text-red-400' : 'bg-[#1a1a20] border-white/5 text-gray-300'
            }`}>
              {isAnomaly ? '⚠️' : '⚡'} {insight}
            </div>
          );
        })}
      </div>

      {/* Main Timeline Axis */}
      <div className="relative h-24 mt-12 mb-8">
        <div className="absolute top-1/2 left-0 w-full h-[1px] bg-white/10" />
        
        {timeline.events.map((event, idx) => {
          const pos = ((new Date(event.timestamp).getTime() - startTime) / duration) * 100;
          const isEven = idx % 2 === 0;

          return (
            <div 
              key={idx}
              className="absolute group cursor-pointer"
              style={{ left: `${pos}%`, top: '50%', transform: 'translate(-50%, -50%)' }}
              onMouseEnter={() => setSelectedEvent(event)}
            >
              {/* Event Dot */}
              <div className={`w-3 h-3 rounded-full border-2 ${EVENT_STYLES[event.type]} group-hover:scale-150 transition-transform relative ${
                event.is_anomaly ? 'shadow-[0_0_8px_rgba(239,68,68,0.8)]' : ''
              }`}>
                {event.is_anomaly && <div className="absolute inset-0 rounded-full animate-ping bg-red-500 opacity-20" />}
              </div>
              {/* Label (Alternating top/bottom) */}
              <div className={`absolute left-1/2 -translate-x-1/2 w-32 text-center transition-all ${
                isEven ? '-top-10' : 'top-6'
              } opacity-40 group-hover:opacity-100`}>
                <p className="text-[10px] font-mono text-white/80 truncate">{new Date(event.timestamp).getFullYear()}</p>
                <p className="text-[9px] text-gray-500 uppercase tracking-tighter truncate">{event.source}</p>
              </div>

              {/* Selection Indicator */}
              {selectedEvent === event && (
                <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-8 h-8 rounded-full bg-[#00ffcc]/10 animate-ping" />
              )}
            </div>
          );
        })}
      </div>

      {/* Detail Tooltip Section */}
      <div className="min-h-[60px] border-t border-white/5 pt-4">
        {selectedEvent ? (
          <div className="flex items-start justify-between">
            <div>
              <div className="flex items-center gap-2">
                <p className="text-xs text-[#00ffcc] font-mono">{new Date(selectedEvent.timestamp).toUTCString()}</p>
                {selectedEvent.is_anomaly && <span className="text-[9px] px-1 bg-red-500/20 text-red-400 border border-red-500/30 rounded uppercase font-bold">Anomaly Detected</span>}
              </div>
              <p className="text-sm text-gray-200 mt-1">{selectedEvent.description}</p>
            </div>
            <div className="text-right">
              <span className="text-[10px] px-2 py-0.5 rounded bg-white/5 border border-white/10 text-gray-400 uppercase">
                {selectedEvent.source}
              </span>
            </div>
          </div>
        ) : (
          <p className="text-xs text-gray-600 italic text-center py-2">Hover over an event marker to inspect temporal data</p>
        )}
      </div>
    </div>
  );
};

export default TimelineView;