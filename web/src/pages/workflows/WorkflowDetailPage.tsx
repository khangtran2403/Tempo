
import React, { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Edit, Play, Trash2, RefreshCw, History, Clock } from 'lucide-react';
import toast from 'react-hot-toast';
import { workflowsAPI } from '../../api/workflows';
import { executionsAPI } from '../../api/executions';
import WebhookHistoryList from '../../component/workflow/WebhookHistoryList';
import { ExecutionListResponse, ExecutionResponse } from '../../types/execution'; // Import new types

type Tab = 'executions' | 'webhooks';

export default function WorkflowDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [activeTab, setActiveTab] = useState<Tab>('executions');

  const { data: workflow, isLoading } = useQuery({
    queryKey: ['workflow', id],
    queryFn: () => workflowsAPI.getById(id!),
    enabled: !!id,
  });

  const { data: executionsData, refetch: refetchExecutions } = useQuery<ExecutionListResponse>({ // Use ExecutionListResponse type
    queryKey: ['executions', id],
    queryFn: () => executionsAPI.listByWorkflowId(id!), // Use the new listByWorkflowId function
    refetchInterval: autoRefresh ? 5000 : false,
    enabled: !!id && activeTab === 'executions',
  });

  const triggerMutation = useMutation({
    mutationFn: () => workflowsAPI.trigger(id!),
    onSuccess: () => {
      toast.success('Dã kích hoạt workflow!');
      setTimeout(() => refetchExecutions(), 1000);
    },
    onError: (err: any) => {
      toast.error(`Kích hoạt workflow thất bại: ${err.message}`);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => workflowsAPI.delete(id!),
    onSuccess: () => {
      toast.success('Đã xóa workflow');
      navigate('/workflows');
    },
    onError: (err: any) => {
      toast.error(`Xóa workflow thất bại: ${err.message}`);
    },
  });

  if (!id) {
    navigate('/workflows');
    return null;
  }
  
  if (isLoading) {
    return <div className="flex items-center justify-center h-full"><div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div></div>;
  }

  if (!workflow) {
    return <div className="text-center py-12 text-gray-600">Không tìm thấy workflow nào</div>;
  }

  const executions = executionsData?.executions || []; // Correctly access the executions array

  return (
    <div className="space-y-6">
      <div className="card">
        <div className="flex items-start justify-between mb-4">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{workflow.name}</h1>
            <p className="text-gray-600 mt-2">{workflow.description}</p>
          </div>
          <div className="flex items-center space-x-2">
            <button onClick={() => navigate(`/workflows/${id}/edit`)} className="btn btn-secondary flex items-center"><Edit className="w-4 h-4 mr-2" />Chỉnh sửa</button>
            <button onClick={() => triggerMutation.mutate()} disabled={!workflow.is_active || triggerMutation.isPending} className="btn btn-primary flex items-center"><Play className="w-4 h-4 mr-2" />Kích hoạt</button>
            <button onClick={() => { if (window.confirm('Bạn có muốn xóa?')) { deleteMutation.mutate(); } }} className="btn btn-danger"><Trash2 className="w-4 h-4" /></button>
          </div>
        </div>
      </div>

      <div className="card">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-6">
            <button onClick={() => setActiveTab('executions')} className={`shrink-0 border-b-2 px-1 pb-4 text-sm font-medium flex items-center ${activeTab === 'executions' ? 'border-primary-500 text-primary-600' : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}`}>
              <Clock className="w-4 h-4 mr-2" />
             Lịch sử thực thi
            </button>
            <button onClick={() => setActiveTab('webhooks')} className={`shrink-0 border-b-2 px-1 pb-4 text-sm font-medium flex items-center ${activeTab === 'webhooks' ? 'border-primary-500 text-primary-600' : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}`}>
              <History className="w-4 h-4 mr-2" />
               Lịch sử gọi Webhook 
            </button>
          </nav>
        </div>

        <div className="mt-6">
          {activeTab === 'executions' && (
            <div>
              <div className="flex items-center justify-end mb-4">
                <button onClick={() => setAutoRefresh(!autoRefresh)} className={`p-2 rounded-lg ${autoRefresh ? 'bg-green-100' : 'bg-gray-100'}`} title="Auto-refresh">
                  <RefreshCw className={`w-4 h-4 ${autoRefresh ? 'text-green-600 animate-spin' : 'text-gray-600'}`} />
                </button>
              </div>
              {executions.length === 0 ? (
                <div className="text-center py-8 text-gray-600">Chưa có lần thực thi nào cho workflow này.</div>
              ) : (
                <div className="space-y-2">
                  {executions.map((exec: ExecutionResponse) => ( // Use ExecutionResponse type for map
                    <ExecutionLogItem key={exec.id} execution={exec} />
                  ))}
                </div>
              )}
            </div>
          )}

          {activeTab === 'webhooks' && <WebhookHistoryList workflowId={id} />}
        </div>
      </div>
    </div>
  );
}

function ExecutionLogItem({ execution }: { execution: ExecutionResponse }) { // Use ExecutionResponse type
  const statusColors: Record<string, string> = {
    success: 'bg-green-100 text-green-700',
    failed: 'bg-red-100 text-red-700',
    running: 'bg-yellow-100 text-yellow-700',
    timeout: 'bg-gray-100 text-gray-700',
  };

  return (
    <Link to={`/executions/${execution.id}`} className="block p-3 border border-gray-200 rounded-lg hover:border-primary-300 hover:bg-primary-50 transition-all">
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <div className="flex items-center space-x-3">
            <span className={`px-2 py-1 rounded text-xs font-medium ${statusColors[execution.status.toLowerCase()]}`}>{/* Fix: toLowerCase() */}
              {execution.status.toUpperCase()}
            </span>
            <span className="text-sm text-gray-600">
              {new Date(execution.started_at).toLocaleString()}
            </span>
          </div>
          {execution.error_message && (
            <p className="text-sm text-red-600 mt-1">{execution.error_message}</p>
          )}
        </div>
        <div className="text-right">
          <p className="text-sm text-gray-600">
            {execution.duration_ms ? `${execution.duration_ms}ms` : '-'}
          </p>
        </div>
      </div>
    </Link>
  );
}
