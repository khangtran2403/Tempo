
import React from 'react';
import { Zap, Globe, Mail, Clock, Save, GitBranch, Bot, Table, Sheet, UploadCloud, FolderArchive, BookText } from 'lucide-react';

interface NodeSidebarProps {
  onAddNode: (type: string, nodeType: 'trigger' | 'action') => void;
  onSave: () => void;
}

const triggers = [
  { type: 'webhook', label: 'Lời gọi Webhook', icon: Zap },
  { type: 'cron', label: 'Lên lịch', icon: Clock },
];

const actions = [
  { type: 'http', label: 'Gửi yêu cầu HTTP ', icon: Globe },
  { type: 'email', label: 'Gửi Email', icon: Mail },
  { type: 'github', label: 'Github', icon: GitBranch },
  { type: 'discord', label: 'Gửi tin nhắn', icon: Bot },
  { type: 'excel', label: 'Xuất ra file Excel', icon: Table },
  { type: 'google_sheets', label: 'Google Sheets', icon: Sheet },
  { type: 'gcs', label: 'Tải file lên Google Cloud', icon: UploadCloud },
  { type: 'google_drive', label: 'Tải file lên Google Drive', icon: FolderArchive },
  { type: 'notion', label: 'Notion', icon: BookText },
];


export default function NodeSidebar({ onAddNode, onSave }: NodeSidebarProps) {
  return (
    <div className="w-64 bg-white border-l border-gray-200 p-4 space-y-6">
      <div>
        <h3 className="text-sm font-semibold text-gray-700 mb-3 uppercase">Triggers</h3>
        <div className="space-y-2">
          {triggers.map((trigger) => (
            <button
              key={trigger.type}
              onClick={() => onAddNode(trigger.type, 'trigger')}
              className="w-full flex items-center space-x-3 px-3 py-2 border border-gray-200 rounded-lg hover:border-yellow-500 hover:bg-yellow-50 transition-all"
            >
              <trigger.icon className="w-5 h-5 text-yellow-600" />
              <span className="text-sm font-medium text-gray-700">{trigger.label}</span>
            </button>
          ))}
        </div>
      </div>

      <div>
        <h3 className="text-sm font-semibold text-gray-700 mb-3 uppercase">Actions</h3>
        <div className="space-y-2">
          {actions.map((action) => (
            <button
              key={action.type}
              onClick={() => onAddNode(action.type, 'action')}
              className="w-full flex items-center space-x-3 px-3 py-2 border border-gray-200 rounded-lg hover:border-primary-500 hover:bg-primary-50 transition-all"
            >
              <action.icon className="w-5 h-5 text-primary-600" />
              <span className="text-sm font-medium text-gray-700">{action.label}</span>
            </button>
          ))}
        </div>
      </div>

      <button
        onClick={onSave}
        className="w-full btn btn-primary flex items-center justify-center space-x-2"
      >
        <Save className="w-5 h-5" />
        <span>Lưu Workflow</span>
      </button>
    </div>
  );
}