
import React from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { Zap } from 'lucide-react';

export default function TriggerNode({ data }: NodeProps) {
  return (
    <div className="px-4 py-3 shadow-lg rounded-lg bg-gradient-to-r from-yellow-400 to-orange-500 border-2 border-yellow-600 min-w-[200px]">
      <div className="flex items-center space-x-2">
        <Zap className="w-5 h-5 text-white" />
        <div>
          <div className="text-xs text-white font-medium uppercase">Trigger</div>
          <div className="text-sm font-bold text-white">{data.label}</div>
        </div>
      </div>
      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 !bg-white !border-2 !border-yellow-600"
      />
    </div>
  );
}