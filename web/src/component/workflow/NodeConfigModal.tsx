import React, { useEffect, useState } from 'react';
import { Node } from 'reactflow';
import { X } from 'lucide-react';
import toast from 'react-hot-toast';

interface NodeConfigModalProps {
  open: boolean;
  node: Node;
  onClose: () => void;
  onSave: (data: { id: string; config: Record<string, any> }) => void;
}


const GitHubConfig = ({ initialConfig, onConfigChange }: { initialConfig: any, onConfigChange: (config: any) => void }) => {
  const [config, setConfig] = useState(initialConfig);
  const action = config.action || 'create_issue';

  useEffect(() => {
    onConfigChange(config);
  }, [config, onConfigChange]);
  
  const handleActionChange = (newAction: string) => {
    setConfig({
      action: newAction,
      integration_id: config.integration_id,
      repo: config.repo,
    });
  };


  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Action</label>
        <select
          value={action}
          onChange={(e) => handleActionChange(e.target.value)}
          className="input"
        >
          <option value="create_issue">Tạo Issue</option>
          <option value="create_pull_request">Tạo Pull Request</option>
          <option value="add_comment">Thêm bình luận</option>
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Github ID</label>
        <input
          type="text"
          value={config.integration_id || ''}
          onChange={(e) => setConfig({ ...config, integration_id: e.target.value })}
          className="input"
          placeholder="GitHub Integration ID"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Repository (e.g., owner/repo-name)</label>
        <input
          type="text"
          value={config.repo || ''}
          onChange={(e) => setConfig({ ...config, repo: e.target.value })}
          className="input"
          placeholder="owner/repo-name"
        />
      </div>

      {action === 'create_issue' && (
        <>
          <input type="text" value={config.title || ''} onChange={(e) => setConfig({ ...config, title: e.target.value })} className="input" placeholder="Issue Title" />
          <textarea value={config.body || ''} onChange={(e) => setConfig({ ...config, body: e.target.value })} className="input" placeholder="Issue Body" />
          <input type="text" value={(config.labels || []).join(', ')} onChange={(e) => setConfig({ ...config, labels: e.target.value.split(',').map(s => s.trim()) })} className="input" placeholder="Labels (comma-separated)" />
        </>
      )}

      {action === 'create_pull_request' && (
        <>
          <input type="text" value={config.title || ''} onChange={(e) => setConfig({ ...config, title: e.target.value })} className="input" placeholder="PR Title" />
          <input type="text" value={config.head || ''} onChange={(e) => setConfig({ ...config, head: e.target.value })} className="input" placeholder="Head Branch (your branch)" />
          <input type="text" value={config.base || ''} onChange={(e) => setConfig({ ...config, base: e.target.value })} className="input" placeholder="Base Branch (e.g., main)" />
          <textarea value={config.body || ''} onChange={(e) => setConfig({ ...config, body: e.target.value })} className="input" placeholder="PR Body" />
        </>
      )}

      {action === 'add_comment' && (
        <>
          <input type="text" value={config.issue_number || ''} onChange={(e) => setConfig({ ...config, issue_number: e.target.value })} className="input" placeholder="Issue/PR Number" />
          <textarea value={config.text || ''} onChange={(e) => setConfig({ ...config, text: e.target.value })} className="input" placeholder="Comment Text" />
        </>
      )}
    </div>
  );
};



const DiscordConfig = ({ initialConfig, onConfigChange }: { initialConfig: any, onConfigChange: (config: any) => void }) => {
  const [config, setConfig] = useState(initialConfig);
  const action = config.action || 'send_message';

  useEffect(() => {
    onConfigChange(config);
  }, [config, onConfigChange]);

  const handleActionChange = (newAction: string) => {
    setConfig({ action: newAction, webhook_url: config.webhook_url });
  };

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Action</label>
        <select
          value={action}
          onChange={(e) => handleActionChange(e.target.value)}
          className="input"
        >
          <option value="send_message">Gửi tin nhắn</option>
          <option value="send_embed">Send Embed</option>
        </select>
      </div>

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Đường dẫn Discord Webhook</label>
        <input
          type="url"
          value={config.webhook_url || ''}
          onChange={(e) => setConfig({ ...config, webhook_url: e.target.value })}
          className="input"
          placeholder="https://discord.com/api/webhooks/..."
        />
      </div>

      {action === 'send_message' && (
        <>
          <textarea value={config.content || ''} onChange={(e) => setConfig({ ...config, content: e.target.value })} className="input" placeholder="Message Content" />
          <input type="text" value={config.username || ''} onChange={(e) => setConfig({ ...config, username: e.target.value })} className="input" placeholder="Bot Username (optional)" />
          <input type="url" value={config.avatar_url || ''} onChange={(e) => setConfig({ ...config, avatar_url: e.target.value })} className="input" placeholder="Bot Avatar URL (optional)" />
        </>
      )}

      {action === 'send_embed' && (
        <>
          <input type="text" value={config.title || ''} onChange={(e) => setConfig({ ...config, title: e.target.value })} className="input" placeholder="Embed Title" />
          <textarea value={config.description || ''} onChange={(e) => setConfig({ ...config, description: e.target.value })} className="input" placeholder="Embed Description" />
          <input type="number" value={config.color || 0} onChange={(e) => setConfig({ ...config, color: parseInt(e.target.value, 10) })} className="input" placeholder="Embed Color (decimal)" />
        </>
      )}
    </div>
  );
};


export default function NodeConfigModal({ open, node, onSave, onClose }: NodeConfigModalProps) {
  const [config, setConfig] = useState(node.data.config || {});
  const [editableId, setEditableId] = useState(node.id);
  
  const [rawHeaders, setRawHeaders] = useState(JSON.stringify(node.data.config.headers || {}, null, 2));
  const [rawBody, setRawBody] = useState(JSON.stringify(node.data.config.body || {}, null, 2));
  const [rawProperties, setRawProperties] = useState(JSON.stringify(node.data.config.properties || {}, null, 2));

  useEffect(() => {
    const currentConfig = node.data.config || {};
    setConfig(currentConfig);
    setEditableId(node.id);
    setRawHeaders(JSON.stringify(currentConfig.headers || {}, null, 2));
    setRawBody(JSON.stringify(currentConfig.body || {}, null, 2));
    setRawProperties(JSON.stringify(currentConfig.properties || {}, null, 2));
  }, [node]);

  if (!open) return null;

  const handleSave = () => {
    if (!editableId || editableId.trim() === '' || editableId.includes(' ')) {
      toast.error('Node ID không hợp lệ. Không được để trống và không được chứa khoảng trắng.');
      return;
    }

    let finalConfig = { ...config };
    if (node.data.label === 'http') {
      try {
        finalConfig.headers = rawHeaders ? JSON.parse(rawHeaders) : {};
        finalConfig.body = rawBody ? JSON.parse(rawBody) : {};
      } catch (e) {
        toast.error("HTTP Headers hoặc Body không phải là JSON hợp lệ.");
        return;
      }
    }
    if (node.data.label === 'notion') {
      finalConfig.properties = rawProperties;
    }
    
    onSave({ id: editableId, config: finalConfig });
  };

  const renderConfigFields = () => {
    switch (node.data.label) {
      case 'webhook':
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Webhook Token (tùy chọn)
            </label>
            <input
              type="text"
              value={config.webhook_token || ''}
              onChange={(e) => setConfig({ ...config, webhook_token: e.target.value })}
              className="input"
              placeholder="Nhập token webhook cho việc xác thực"
            />
          </div>
        );

      case 'cron':
        return (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Lên lịch
            </label>
            <input
              type="text"
              value={config.expression || ''}
              onChange={(e) => setConfig({ ...config, expression: e.target.value })}
              className="input"
              placeholder="0 9 * * * (runs at 9 AM daily)"
            />
            <p className="text-xs text-gray-500 mt-2">Format: minute hour day-of-month month day-of-week</p>
          </div>
        );

      case 'http':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Phương thức
              </label>
              <select
                value={config.method || 'GET'}
                onChange={(e) => setConfig({ ...config, method: e.target.value })}
                className="input"
              >
                <option>GET</option>
                <option>POST</option>
                <option>PUT</option>
                <option>PATCH</option>
                <option>DELETE</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Đường dẫn URL
              </label>
              <input
                type="url"
                value={config.url || ''}
                onChange={(e) => setConfig({ ...config, url: e.target.value })}
                className="input"
                placeholder="https://api.example.com/endpoint"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Headers (JSON)
              </label>
              <textarea
                value={rawHeaders}
                onChange={(e) => setRawHeaders(e.target.value)}
                className="input font-mono text-xs"
                rows={5}
                placeholder='{ "Content-Type": "application/json" }'
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Body (JSON)
              </label>
              <textarea
                value={rawBody}
                onChange={(e) => setRawBody(e.target.value)}
                className="input font-mono text-xs"
                rows={5}
                placeholder='{ "key": "{{ .trigger.data.value }}" }'
              />
            </div>
          </div>
        );

      case 'email':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tới Email</label>
              <input type="email" value={config.to || ''} onChange={(e) => setConfig({ ...config, to: e.target.value })} className="input" placeholder="recipient@example.com" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tiêu đề</label>
              <input type="text" value={config.subject || ''} onChange={(e) => setConfig({ ...config, subject: e.target.value })} className="input" placeholder="Email Subject" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Nội dung</label>
              <textarea value={config.body || ''} onChange={(e) => setConfig({ ...config, body: e.target.value })} className="input" rows={4} placeholder="Email body content..." />
            </div>
          </div>
        );

      case 'github':
        return <GitHubConfig initialConfig={config} onConfigChange={setConfig} />;
        
      case 'discord':
        return <DiscordConfig initialConfig={config} onConfigChange={setConfig} />;

      case 'excel':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Tên file (Filename)
              </label>
              <input
                type="text"
                value={config.filename || ''}
                onChange={(e) => setConfig({ ...config, filename: e.target.value })}
                className="input"
                placeholder="report-{{ .trigger.data.date }}.xlsx"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Dữ liệu (Data)
              </label>
              <input
                type="text"
                value={config.data || ''}
                onChange={(e) => setConfig({ ...config, data: e.target.value })}
                className="input"
                placeholder="{{ .get_users.body.users }}"
              />
              <p className="text-xs text-gray-500 mt-1">Sử dụng biến template để trỏ đến một mảng (array) dữ liệu từ action trước đó.</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Tiêu đề cột (Headers)
              </label>
              <input
                type="text"
                value={(config.headers || []).join(', ')}
                onChange={(e) => setConfig({ ...config, headers: e.target.value.split(',').map(s => s.trim())})}
                className="input"
                placeholder="id, name, email, created_at"
              />
              <p className="text-xs text-gray-500 mt-1">Các cột, cách nhau bởi dấu phẩy. Nếu bỏ trống, sẽ tự động lấy từ key của object đầu tiên.</p>
            </div>
          </div>
        );

      case 'google_sheets':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ID của tích hợp Google</label>
              <input type="text" value={config.integration_id || ''} onChange={(e) => setConfig({ ...config, integration_id: e.target.value })} className="input" placeholder="ID của tích hợp Google" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ID từ URL của Google Sheet</label>
              <input type="text" value={config.spreadsheet_id || ''} onChange={(e) => setConfig({ ...config, spreadsheet_id: e.target.value })} className="input" placeholder="ID từ URL của Google Sheet" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tên Sheet</label>
              <input type="text" value={config.sheet_name || ''} onChange={(e) => setConfig({ ...config, sheet_name: e.target.value })} className="input" placeholder="Tên của sheet (ví dụ: Sheet1)" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Dữ liệu hàng</label>
              <input type="text" value={config.row_data || ''} onChange={(e) => setConfig({ ...config, row_data: e.target.value })} className="input" placeholder="[{{ .trigger.data.name }}, {{ .trigger.data.email }}]" />
            </div>
          </div>
        );

      case 'gcs':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ID tích hợp Google</label>
              <input type="text" value={config.integration_id || ''} onChange={(e) => setConfig({ ...config, integration_id: e.target.value })} className="input" placeholder="ID của tích hợp Google" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tên Bucket</label>
              <input type="text" value={config.bucket_name || ''} onChange={(e) => setConfig({ ...config, bucket_name: e.target.value })} className="input" placeholder="Tên Google Cloud Storage bucket" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tên file</label>
              <input type="text" value={config.object_name || ''} onChange={(e) => setConfig({ ...config, object_name: e.target.value })} className="input" placeholder="path/to/your/file.txt" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Đường dẫn file (tùy chọn)</label>
              <input type="text" value={config.file_path || ''} onChange={(e) => setConfig({ ...config, file_path: e.target.value })} className="input" placeholder="{{ .excel_action.file_path }}" />
              <p className="text-xs text-gray-500 mt-1">Đường dẫn đến file cục bộ để tải lên (ví dụ: từ action Excel).</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Nội dung file (nếu không dùng File Path)</label>
              <textarea value={config.content || ''} onChange={(e) => setConfig({ ...config, content: e.target.value })} className="input" rows={4} placeholder="Nội dung file..." />
            </div>
          </div>
        );

      case 'google_drive':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ID tích hợp Google</label>
              <input type="text" value={config.integration_id || ''} onChange={(e) => setConfig({ ...config, integration_id: e.target.value })} className="input" placeholder="ID của tích hợp Google" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Tên file</label>
              <input type="text" value={config.filename || ''} onChange={(e) => setConfig({ ...config, filename: e.target.value })} className="input" placeholder="my-document.txt" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ID thư mục (tùy chọn)</label>
              <input type="text" value={config.parent_folder_id || ''} onChange={(e) => setConfig({ ...config, parent_folder_id: e.target.value })} className="input" placeholder="ID của thư mục trên Google Drive" />
            </div>
             <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Content Type (MIME Type)</label>
              <input type="text" value={config.content_type || ''} onChange={(e) => setConfig({ ...config, content_type: e.target.value })} className="input" placeholder="text/plain" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Đường dẫn file (tùy chọn)</label>
              <input type="text" value={config.file_path || ''} onChange={(e) => setConfig({ ...config, file_path: e.target.value })} className="input" placeholder="{{ .excel_action.file_path }}" />
              <p className="text-xs text-gray-500 mt-1">Đường dẫn đến file cục bộ để tải lên.</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Nội dung file (nếu không dùng File Path)</label>
              <textarea value={config.content || ''} onChange={(e) => setConfig({ ...config, content: e.target.value })} className="input" rows={4} placeholder="Nội dung file..." />
            </div>
          </div>
        );

      case 'notion':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">ID tích hợp Notion</label>
              <input type="text" value={config.integration_id || ''} onChange={(e) => setConfig({ ...config, integration_id: e.target.value })} className="input" placeholder="ID của tích hợp Notion" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Database ID</label>
              <input type="text" value={config.database_id || ''} onChange={(e) => setConfig({ ...config, database_id: e.target.value })} className="input" placeholder="ID của Notion database" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Thuộc tính</label>
              <textarea
                value={rawProperties}
                onChange={(e) => setRawProperties(e.target.value)}
                className="input font-mono text-xs"
                rows={10}
                placeholder={`{
  "Name": {
    "title": [
      {
        "text": {
          "content": "New Lead: {{ .trigger.data.name }}"
        }
      }
    ]
  }
}`}
              />
               <p className="text-xs text-gray-500 mt-1">Nhập theo cấu trúc trên.</p>
            </div>
             <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Nội dung trang</label>
              <textarea value={config.content || ''} onChange={(e) => setConfig({ ...config, content: e.target.value })} className="input" rows={5} placeholder="Nội dung trang..." />
            </div>
          </div>
        );

      default:
        return <div className="text-gray-600">Không có cấu hình nào cho nút này.</div>;
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-lg max-w-2xl w-full mx-4">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-gray-900">
            Cấu hình {node.data.label}
          </h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Config form */}
        <div className="px-6 py-4 space-y-4 max-h-[60vh] overflow-y-auto">
          {renderConfigFields()}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end space-x-3 border-t border-gray-200 px-6 py-4">
          <button onClick={onClose} className="btn btn-secondary">
            Hủy
          </button>
          <button onClick={handleSave} className="btn btn-primary">
            Lưu cấu hình
          </button>
        </div>
      </div>
    </div>
  );
}