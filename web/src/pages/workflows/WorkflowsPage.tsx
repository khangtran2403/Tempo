
import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { Plus, Search, Trash2, Edit, Play } from 'lucide-react';
import toast from 'react-hot-toast';
import { workflowsAPI } from '../../api/workflows';
import { Workflow } from '../../types/workflow';

export default function WorkflowsPage() {
  const [searchQuery, setSearchQuery] = useState('');
  const queryClient = useQueryClient();

  
  const { data, isLoading } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => workflowsAPI.list(1, 100),
  });


  const deleteMutation = useMutation({
    mutationFn: (id: string) => workflowsAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workflows'] });
      toast.success('Đã xóa workflow');
    },
    onError: () => {
      toast.error('Xóa workflow thất bại');
    },
  });

  
  const triggerMutation = useMutation({
    mutationFn: (id: string) => workflowsAPI.trigger(id),
    onSuccess: () => {
      toast.success('Kích hoạt workflow thành công');
    },
    onError: () => {
      toast.error('Kích hoạt workflow thất bại');
    },
  });

  const workflows: Workflow[] = data?.workflows || [];

  
  const filteredWorkflows = workflows.filter((wf) =>
    wf.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    wf.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Workflows</h1>
          <p className="text-gray-600 mt-1">Quản lý workflows của bạn</p>
        </div>
        <Link to="/workflows/new" className="btn btn-primary flex items-center space-x-2">
          <Plus className="w-5 h-5" />
          <span>Workflow mới</span>
        </Link>
      </div>

      {/* Search bar */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
        <input
          type="text"
          placeholder="Tìm workflows..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
        />
      </div>

      {/* Workflows grid */}
      {isLoading ? (
        <div className="text-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
          <p className="text-gray-600 mt-4">Loading workflows...</p>
        </div>
      ) : filteredWorkflows.length === 0 ? (
        <div className="text-center py-12 card">
          <p className="text-gray-600">Không có workflow nào được tìm thấy</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredWorkflows.map((workflow) => (
            <WorkflowCard
              key={workflow.id}
              workflow={workflow}
              onDelete={() => {
                if (window.confirm('Bạn có chắc muốn xóa workflow này không?')) {
                  deleteMutation.mutate(workflow.id);
                }
              }}
              onTrigger={() => triggerMutation.mutate(workflow.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}


interface WorkflowCardProps {
  workflow: Workflow;
  onDelete: () => void;
  onTrigger: () => void;
}

function WorkflowCard({ workflow, onDelete, onTrigger }: WorkflowCardProps) {
  return (
    <div className="card hover:shadow-lg transition-shadow">
      <div className="flex items-start justify-between mb-4">
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-gray-900 mb-1">
            {workflow.name}
          </h3>
          <p className="text-sm text-gray-600 line-clamp-2">
            {workflow.description || 'Không có mô tả nào'}
          </p>
        </div>
        <span className={`px-2 py-1 rounded-full text-xs font-medium ${
          workflow.is_active
            ? 'bg-green-100 text-green-700'
            : 'bg-gray-100 text-gray-600'
        }`}>
          {workflow.is_active ? 'Hoạt động' : 'Không hoạt động'}
        </span>
      </div>

      {/* Actions count */}
      <div className="flex items-center space-x-4 text-sm text-gray-600 mb-4">
        <span>{workflow.definition.actions.length} actions</span>
        <span>•</span>
        <span>v{workflow.version}</span>
      </div>

      {/* Buttons */}
      <div className="flex items-center space-x-2">
        <Link
          to={`/workflows/${workflow.id}`}
          className="flex-1 btn btn-secondary text-center text-sm"
        >
          Xem
        </Link>
        <Link
          to={`/executions?workflowId=${workflow.id}`} 
          className="flex-1 btn btn-secondary text-center text-sm"
        >
          Lịch sử thực thi
        </Link>
        <Link
          to={`/workflows/${workflow.id}/edit`}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          title="Edit"
        >
          <Edit className="w-4 h-4 text-gray-600" />
        </Link>
        <button
          onClick={onTrigger}
          className="p-2 hover:bg-green-100 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          title="Trigger"
          disabled={!workflow.is_active}
        >
          <Play className="w-4 h-4 text-green-600" />
        </button>
        <button
          onClick={onDelete}
          className="p-2 hover:bg-red-100 rounded-lg transition-colors"
          title="Delete"
        >
          <Trash2 className="w-4 h-4 text-red-600" />
        </button>
      </div>
    </div>
  );
}