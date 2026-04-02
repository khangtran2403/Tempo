
import React from 'react';
import { NavLink } from 'react-router-dom';
import { 
  LayoutDashboard, 
  Workflow, 
  Settings,
  LockKeyhole,
  Share2,
  BookOpen
} from 'lucide-react';

const navItems = [
  { to: '/dashboard', icon: LayoutDashboard, label: 'Bàn làm việc' },
  { to: '/workflows', icon: Workflow, label: 'Workflows' },
  { to: '/secrets', icon: LockKeyhole , label: 'Quản lý các khóa bảo mật' },
  { to: '/integrations', icon: Share2, label: 'Tích hợp' },
  { to: '/documentation', icon: BookOpen, label: 'Tài liệu' },
  { to: '/settings', icon: Settings, label: 'Cài đặt' },
];

export default function Sidebar() {
  return (
    <aside className="w-64 bg-lime-500 border-r border-gray-200 h-full">
      <nav className="p-4 space-y-1">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              `flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors ${
                isActive
                  ? 'bg-primary-50 text-primary-700'
                  : 'text-gray-700 hover:bg-gray-50'
              }`
            }
          >
            <item.icon className="w-5 h-5" />
            <span className="font-medium">{item.label}</span>
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}