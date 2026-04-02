import React from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query'; // Import useQuery
import { executionsAPI } from '../../api/executions';
import { ExecutionResponse, ExecutionListResponse } from '../../types/execution'; // Import types

export default function ExecutionsListPage() {
  const [searchParams] = useSearchParams();
  const workflowId = searchParams.get('workflowId');

  const { data, isLoading, error } = useQuery<ExecutionListResponse | ExecutionResponse[]>({
    queryKey: ['executions', workflowId],
    queryFn: () => {
      if (workflowId) {
        return executionsAPI.listByWorkflowId(workflowId);
      }
      return executionsAPI.getAll();
    },
  });

  const executions: ExecutionResponse[] = [];
  if (data) {
    if (workflowId) {
      // If workflowId is present, data is ExecutionListResponse
      executions.push(...(data as ExecutionListResponse).executions);
    } else {
      // If no workflowId, data is ExecutionResponse[]
      executions.push(...(data as ExecutionResponse[]));
    }
  }

  if (isLoading) return <div className="p-6">Loading...</div>;
  if (error) return <div className="p-6 text-red-600">Lỗi: {error.message}</div>;

  const pageTitle = workflowId ? `Lịch sử thực thi cho Workflow: ${workflowId}` : 'Lịch sử thực thi';

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-semibold">{pageTitle}</h1>
      </div>

      <div className="bg-white shadow rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Workflow</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Trạng thái</th>
              <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Đã bắt đầu</th>
              <th className="px-4 py-2"></th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {executions.map((ex) => (
              <tr key={ex.id}>
                <td className="px-4 py-2 font-mono text-sm">{ex.id}</td>
                <td className="px-4 py-2">{ex.workflow_id}</td>
                <td className="px-4 py-2"><span className="px-2 py-1 rounded text-xs bg-gray-100">{ex.status}</span></td>
                <td className="px-4 py-2">{new Date(ex.started_at).toLocaleString()}</td>
                <td className="px-4 py-2 text-right">
                  <Link className="text-primary-600 hover:text-primary-700" to={`/executions/${ex.id}`}>Xem</Link>
                </td>
              </tr>
            ))}
            {executions.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-6 text-center text-gray-500">Chưa có executions nào</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
