
import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { 
  Workflow, 
  CheckCircle, 
  Clock,
  Plus,
  TrendingUp
} from 'lucide-react';
import { workflowsAPI } from '../../api/workflows';
import { Workflow as WorkflowType } from '../../types/workflow';

export default function DashboardPage() {
  
  const { data: workflowsData, isLoading } = useQuery({
    queryKey: ['workflows'],
    queryFn: () => workflowsAPI.list(1, 10),
  });

  const workflows: WorkflowType[] = workflowsData?.workflows || [];
  const activeWorkflows = workflows.filter(w => w.is_active).length;

  return (
    <div className="space-y-6">
      
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Bàn làm việc</h1>
          <p className="text-gray-600 mt-1">Chào mừng trở lại! Đây là phần tổng quan của bạn.</p>
        </div>
        <Link to="/workflows/new" className="btn btn-primary flex items-center space-x-2">
          <Plus className="w-5 h-5" />
          <span>Tạo Workflow mới</span>
        </Link>
      </div>

    
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <StatCard
          icon={Workflow}
          label="Tất cả workflows"
          value={workflows.length}
          color="blue"
        />
        <StatCard
          icon={CheckCircle}
          label="Đang hoạt động"
          value={activeWorkflows}
          color="green"
        />
        <StatCard
          icon={Clock}
          label="Đang chạy"
          value={0}
          color="yellow"
        />
      </div>

      
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">Những Workflows gần đây</h2>
          <Link to="/workflows" className="text-primary-600 hover:text-primary-700 text-sm font-medium">
            Xem tất cả workflow
          </Link>
        </div>

        {isLoading ? (
          <div className="text-center py-8 text-gray-500">Loading...</div>
        ) : workflows.length === 0 ? (
          <div className="text-center py-8">
            <Workflow className="w-12 h-12 text-gray-400 mx-auto mb-3" />
            <p className="text-gray-600 mb-4">Chưa có workflow nào</p>
            <Link to="/workflows/new" className="btn btn-primary">
              Tạo Workflow đầu tiên
            </Link>
          </div>
        ) : (
          <div className="space-y-3">
            {workflows.slice(0, 5).map((workflow) => (
              <WorkflowItem key={workflow.id} workflow={workflow} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}


interface StatCardProps {
  icon: React.ElementType;
  label: string;
  value: string | number;
  color: 'blue' | 'green' | 'yellow' | 'purple';
}

function StatCard({ icon: Icon, label, value, color }: StatCardProps) {
  const colorClasses = {
    blue: 'bg-blue-100 text-blue-600',
    green: 'bg-green-100 text-green-600',
    yellow: 'bg-yellow-100 text-yellow-600',
    purple: 'bg-purple-100 text-purple-600',
  };

  return (
    <div className="card">
      <div className="flex items-center space-x-4">
        <div className={`p-3 rounded-lg ${colorClasses[color]}`}>
          <Icon className="w-6 h-6" />
        </div>
        <div>
          <p className="text-sm text-gray-600">{label}</p>
          <p className="text-2xl font-bold text-gray-900">{value}</p>
        </div>
      </div>
    </div>
  );
}


function WorkflowItem({ workflow }: { workflow: WorkflowType }) {
  return (
    <Link
      to={`/workflows/${workflow.id}`}
      className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:border-primary-300 hover:bg-primary-50 transition-all"
    >
      <div className="flex items-center space-x-4">
        <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
          workflow.is_active ? 'bg-green-100' : 'bg-gray-100'
        }`}>
          <Workflow className={`w-5 h-5 ${
            workflow.is_active ? 'text-green-600' : 'text-gray-400'
          }`} />
        </div>
        <div>
          <h3 className="font-medium text-gray-900">{workflow.name}</h3>
          <p className="text-sm text-gray-500">{workflow.description || 'Không có mô tả'}</p>
        </div>
      </div>
      <div className="flex items-center space-x-4">
        <span className={`px-3 py-1 rounded-full text-xs font-medium ${
          workflow.is_active
            ? 'bg-green-100 text-green-700'
            : 'bg-gray-100 text-gray-600'
        }`}>
          {workflow.is_active ? 'Hoạt động' : 'Không hoạt động'}
        </span>
      </div>
    </Link>
  );
}