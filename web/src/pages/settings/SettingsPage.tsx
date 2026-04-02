import React, { useState } from 'react';
import toast from 'react-hot-toast';

export default function SettingsPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [saving, setSaving] = useState(false);

  const onSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      toast.error('Email không hợp lệ');
      return;
    }
    setSaving(true);
    try {
      // TODO: call settings API to update profile
      await new Promise((r) => setTimeout(r, 600));
      toast.success('Lưu thành công');
    } catch (e: any) {
      toast.error(e?.message || 'Lưu thất bại');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="p-6">
      <h1 className="text-2xl font-semibold mb-6">Cài đặt</h1>

      <div className="card max-w-xl">
        <form onSubmit={onSave} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Tên</label>
            <input className="input" value={name} onChange={(e) => setName(e.target.value)} placeholder="Your name" disabled={saving} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input className="input" type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" disabled={saving} />
          </div>
          <button className="btn btn-primary" disabled={saving}>
            {saving ? 'Đang lưu...' : 'Lưu thay đổi'}
          </button>
        </form>
      </div>
    </div>
  );
}
