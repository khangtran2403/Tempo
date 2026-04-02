import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Plus, Trash2, GitBranch, Bot, MessageCircle, ClipboardCopy } from 'lucide-react';
import { integrationsAPI, Integration } from '../../api/integrations';
import toast from 'react-hot-toast';

const providerIcons: Record<string, React.ElementType> = {
  google: () => <img src="/assets/google-icon.svg" alt="Google" className="w-6 h-6" />,
  notion: () => <img src="/assets/notion-icon.svg" alt="Notion" className="w-6 h-6" />,
  github: GitBranch,
  discord: Bot,
  default: MessageCircle,
};

export default function IntegrationsListPage() {
  const queryClient = useQueryClient();
  const { data: integrations, isLoading, error } = useQuery<Integration[]> ({
    queryKey: ['integrations'],
    queryFn: integrationsAPI.list,
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => integrationsAPI.delete(id),
    onSuccess: () => {
      toast.success('Xóa tích hợp thành công');
      queryClient.invalidateQueries({ queryKey: ['integrations'] });
    },
    onError: (err: any) => {
      toast.error(`Xóa tích hợp thất bại: ${err.message}`);
    },
  });

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast.success('ID đã lưu vào bộ nhớ tạm!');
  };

  if (isLoading) return <div className="p-6">Tải dữ liệu...</div>;
  if (error) return <div className="p-6 text-red-600">Lỗi: {error.message}</div>;

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Các tích hợp</h1>
        <Link to="/integrations/add" className="btn btn-primary flex items-center space-x-2">
          <Plus className="w-5 h-5" />
          <span>Thêm tích hợp</span>
        </Link>
      </div>

      <div className="bg-white shadow rounded-lg">
        <ul className="divide-y divide-gray-200">
          {integrations && integrations.length > 0 ? (
            integrations.map((integration) => {
              const Icon = providerIcons[integration.provider] || providerIcons.default;
              return (
                <li key={integration.id} className="p-4 flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <Icon />
                    <div>
                      <p className="font-medium capitalize">{integration.provider}</p>
                      <p className="text-sm text-gray-500">{integration.account_name}</p>
                      <div className="flex items-center space-x-2 mt-1">
                        <p className="text-xs text-gray-400 font-mono">{integration.id}</p>
                        <button onClick={() => copyToClipboard(integration.id)} title="Copy ID">
                          <ClipboardCopy className="w-3 h-3 text-gray-400 hover:text-gray-600" />
                        </button>
                      </div>
                    </div>
                  </div>
                  <button
                    onClick={() => {
                      if (window.confirm('Bạn có chắc chắn muốn xóa tích hợp này không?')) {
                        deleteMutation.mutate(integration.id);
                      }
                    }}
                    className="btn btn-danger-outline"
                    disabled={deleteMutation.isPending && deleteMutation.variables === integration.id}
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </li>
              );
            })
          ) : (
            <li className="p-6 text-center text-gray-500">Không tìm thấy tích hợp nào.</li>
          )}
        </ul>
      </div>
    </div>
  );
}
