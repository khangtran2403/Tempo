import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import workflowVersionAPI, { CreateWorkflowVersionRequest } from "../../api/workflowversion";

export default function CreateWorkflowVersionPage() {
  const { id } = useParams(); // workflowId
  const navigate = useNavigate();

  const [jsonText, setJsonText] = useState<string>('{\n  "trigger": {},\n  "actions": []\n}');
  const [notes, setNotes] = useState<string>("");

  const submit = async () => {
    let parsed: any;

    try {
      parsed = JSON.parse(jsonText);
    } catch (err) {
      alert("Invalid JSON");
      return;
    }

    const payload: CreateWorkflowVersionRequest = {
      Definition: parsed,
      ChangeSummary: notes
    };

    await workflowVersionAPI.create(id!, payload);
    navigate(`/workflows/${id}/versions`);
  };

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-xl font-semibold">Tạo một phiên bản cho Workflow</h1>

      <div>
        <label className="block font-medium mb-2">Cấu hình</label>
        <textarea
          className="w-full border rounded p-3 font-mono h-64"
          value={jsonText}
          onChange={e => setJsonText(e.target.value)}
        />
      </div>

      <div>
        <label className="block font-medium mb-2">Tóm tắt thay đổi</label>
        <input
          className="w-full border rounded p-2"
          value={notes}
          onChange={e => setNotes(e.target.value)}
        />
      </div>

      <button
        onClick={submit}
        className="px-4 py-2 bg-primary-600 text-white rounded hover:bg-primary-700"
      >
        Tạo phiên bản
      </button>
    </div>
  );
}
