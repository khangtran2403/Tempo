
import React from 'react';
import { Handle, Position, NodeProps } from 'reactflow';
import { Zap, Mail, Globe, Database, GitBranch,Bot } from 'lucide-react';

const iconMap: Record<string, React.ElementType> = {
  http: Globe,
  email: Mail,
  database: Database,
  github : GitBranch,
  discord : Bot,
};

export default function ActionNode({ data }: NodeProps) {
  const Icon = iconMap[data.label] || Zap;

  return (
    <div className="px-4 py-3 shadow-lg rounded-lg bg-white border-2 border-primary-300 hover:border-primary-500 transition-colors min-w-[200px]">
      <Handle
        type="target"
        position={Position.Top}
        className="w-3 h-3 !bg-primary-500"
      />
      
      <div className="flex items-center space-x-2">
        <Icon className="w-5 h-5 text-primary-600" />
        <div>
          <div className="text-xs text-gray-500 font-medium uppercase">Action</div>
          <div className="text-sm font-bold text-gray-900">{data.label}</div>
        </div>
      </div>

      <Handle
        type="source"
        position={Position.Bottom}
        className="w-3 h-3 !bg-primary-500"
      />
    </div>
  );
}