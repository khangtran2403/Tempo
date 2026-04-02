
import React, { useEffect, useState } from 'react';
import { Execution } from '../../types/execution';
import { AlertCircle, CheckCircle, Clock, Loader } from 'lucide-react';

interface ExecutionLogProps {
  execution: Execution;
}

export default function ExecutionLog({ execution }: ExecutionLogProps) {
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({
    input: false,
    output: false,
    error: false,
  });

  const toggleSection = (section: string) => {
    setExpandedSections(prev => ({
      ...prev,
      [section]: !prev[section]
    }));
  };

  const statusIcon = {
    'đang chạy': <Loader className="w-5 h-5 text-yellow-600 animate-spin" />,
    'thành công': <CheckCircle className="w-5 h-5 text-green-600" />,
    'thất bại': <AlertCircle className="w-5 h-5 text-red-600" />,
    'hết thời gian chờ': <Clock className="w-5 h-5 text-gray-600" />,
  }[execution.status];

  return (
    <div className="card border border-gray-200">
      {/* Header */}
      <div className="flex items-center justify-between mb-4 pb-4 border-b border-gray-200">
        <div className="flex items-center space-x-3">
          {statusIcon}
          <div>
            <p className="font-medium text-gray-900">{execution.id}</p>
            <p className="text-sm text-gray-600">
              {new Date(execution.started_at).toLocaleString()}
            </p>
          </div>
        </div>
        {execution.duration_ms && (
          <span className="text-sm font-medium text-gray-700">
            {execution.duration_ms}ms
          </span>
        )}
      </div>

      {/* Sections */}
      <div className="space-y-3">
        {/* Input */}
        {execution.input_data && (
          <Section
            title="Input Data"
            expanded={expandedSections.input}
            onToggle={() => toggleSection('input')}
            content={execution.input_data}
          />
        )}

        {/* Output */}
        {execution.output_data && (
          <Section
            title="Output Data"
            expanded={expandedSections.output}
            onToggle={() => toggleSection('output')}
            content={execution.output_data}
          />
        )}

        {/* Error */}
        {execution.error_message && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3">
            <button
              onClick={() => toggleSection('error')}
              className="w-full flex items-center justify-between text-left font-medium text-red-700 hover:text-red-800"
            >
             Thông báo lỗi
              <span className={`transition-transform ${expandedSections.error ? 'rotate-180' : ''}`}>
                ▼
              </span>
            </button>
            {expandedSections.error && (
              <pre className="mt-2 bg-red-100 p-2 rounded text-xs text-red-900 overflow-x-auto">
                {execution.error_message}
              </pre>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

interface SectionProps {
  title: string;
  expanded: boolean;
  onToggle: () => void;
  content: any;
}

function Section({ title, expanded, onToggle, content }: SectionProps) {
  return (
    <div className="border border-gray-200 rounded-lg overflow-hidden">
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between p-3 hover:bg-gray-50 transition-colors font-medium text-gray-900"
      >
        {title}
        <span className={`transition-transform ${expanded ? 'rotate-180' : ''}`}>
          ▼
        </span>
      </button>
      {expanded && (
        <div className="bg-gray-50 p-4 border-t border-gray-200">
          <pre className="font-mono text-xs text-gray-700 overflow-x-auto">
            {JSON.stringify(content, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}