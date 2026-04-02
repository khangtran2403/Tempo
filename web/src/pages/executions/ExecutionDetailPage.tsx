import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { executionsAPI } from '../../api/executions';
import { ExecutionResponse } from '../../types/execution';
import { ArrowLeft } from 'lucide-react';

export default function ExecutionDetailPage() {
  const { id } = useParams<{ id: string }>();

  const { data: execution, isLoading, error } = useQuery<ExecutionResponse>({
    queryKey: ['execution', id],
    queryFn: () => executionsAPI.getById(id!),
    enabled: !!id,
  });

  if (!id) {
    return <div className="text-center py-12 text-gray-600">Không có ID được cung cấp.</div>;
  }

  if (isLoading) {
    return <div className="flex items-center justify-center h-full"><div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div></div>;
  }

  if (error) {
    return <div className="text-center py-12 text-red-600">Lỗi khi tải lịch sử thục thi: {error.message}</div>;
  }

  if (!execution) {
    return <div className="text-center py-12 text-gray-600">Không tìm thấy lần thực thi nào.</div>;
  }

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Chi tiết về lần thực thi: {execution.id}</h1>
        <Link to={`/workflows/${execution.workflow_id}`} className="flex items-center space-x-2 text-gray-600 hover:text-gray-900">
          <ArrowLeft className="w-5 h-5" />
          <span>Trở về</span>
        </Link>
      </div>

      <div className="bg-white shadow rounded-lg p-6 space-y-4">
        <div>
          <h3 className="text-lg font-medium text-gray-900">Thông tin tổng quan</h3>
          <dl className="mt-2 text-sm text-gray-700">
            <div className="bg-gray-50 px-4 py-2 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt className="font-medium text-gray-500">ID</dt>
              <dd className="mt-1 sm:col-span-2 sm:mt-0 font-mono">{execution.id}</dd>
            </div>
            <div className="bg-white px-4 py-2 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt className="font-medium text-gray-500">Workflow ID</dt>
              <dd className="mt-1 sm:col-span-2 sm:mt-0 font-mono">
                <Link to={`/workflows/${execution.workflow_id}`} className="text-blue-600 hover:underline">
                  {execution.workflow_id}
                </Link>
              </dd>
            </div>
            <div className="bg-gray-50 px-4 py-2 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt className="font-medium text-gray-500">Trạng thái</dt>
              <dd className="mt-1 sm:col-span-2 sm:mt-0">
                <span className={`px-2 py-1 rounded text-xs font-medium ${
                    execution.status === 'thành công' ? 'bg-green-100 text-green-700' :
                    execution.status === 'thất bại' ? 'bg-red-100 text-red-700' :
                    execution.status === 'đang chạy' ? 'bg-yellow-100 text-yellow-700' :
                    'bg-gray-100 text-gray-700'
                  }`}>
                  {execution.status.toUpperCase()}
                </span>
              </dd>
            </div>
            <div className="bg-white px-4 py-2 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
              <dt className="font-medium text-gray-500">Bắt đầu lúc</dt>
              <dd className="mt-1 sm:col-span-2 sm:mt-0">{new Date(execution.started_at).toLocaleString()}</dd>
            </div>
            {execution.completed_at && (
              <div className="bg-gray-50 px-4 py-2 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="font-medium text-gray-500">Hoàn thành lúc</dt>
                <dd className="mt-1 sm:col-span-2 sm:mt-0">{new Date(execution.completed_at).toLocaleString()}</dd>
              </div>
            )}
            {execution.duration_ms && (
              <div className="bg-white px-4 py-2 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="font-medium text-gray-500">Thời lượng</dt>
                <dd className="mt-1 sm:col-span-2 sm:mt-0">{execution.duration_ms} ms</dd>
              </div>
            )}
            {execution.error_message && (
              <div className="bg-gray-50 px-4 py-2 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="font-medium text-gray-500">Lỗi</dt>
                <dd className="mt-1 sm:col-span-2 sm:mt-0 text-red-600">{execution.error_message}</dd>
              </div>
            )}
            {execution.temporal_execution_id && (
              <div className="bg-gray-50 px-4 py-2 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                <dt className="font-medium text-gray-500">Temporal Execution ID</dt>
                <dd className="mt-1 sm:col-span-2 sm:mt-0 font-mono">{execution.temporal_execution_id}</dd>
              </div>
            )}
          </dl>
        </div>

        {execution.input_data && Object.keys(execution.input_data).length > 0 && (
          <div>
            <h3 className="text-lg font-medium text-gray-900 mt-4">Dữ liệu đầu vào</h3>
            <pre className="bg-gray-50 p-3 rounded-md text-sm mt-2">{JSON.stringify(execution.input_data, null, 2)}</pre>
          </div>
        )}

        {execution.output_data && Object.keys(execution.output_data).length > 0 && (
          <div>
            <h3 className="text-lg font-medium text-gray-900 mt-4">Dữ liệu đầu ra</h3>
            <pre className="bg-gray-50 p-3 rounded-md text-sm mt-2">{JSON.stringify(execution.output_data, null, 2)}</pre>
          </div>
        )}
      </div>
    </div>
  );
}
