import { useEffect, useState } from "react";
import { useSearchParams, useParams, Link } from "react-router-dom";
import workflowVersionAPI, { WorkflowVersionDetail } from "../../api/workflowversion";
import DiffViewer from "react-diff-viewer-continued";

export default function WorkflowVersionComparePage() {
  const { id } = useParams();
  const [params] = useSearchParams();

  const baseVersion = params.get("base");
  const [compareVersion, setCompareVersion] = useState<string | null>(null);

  const [baseData, setBaseData] = useState<WorkflowVersionDetail | null>(null);
  const [compareData, setCompareData] = useState<WorkflowVersionDetail | null>(null);

  const [versions, setVersions] = useState<string[]>([]);

  useEffect(() => {
    workflowVersionAPI.list(id!).then(list => {
      setVersions(list.map(v => v.Version).sort());
    });
  }, [id]);

  useEffect(() => {
    if (baseVersion) {
      workflowVersionAPI.get(id!, baseVersion).then(setBaseData);
    }
    if (compareVersion) {
      workflowVersionAPI.get(id!, compareVersion).then(setCompareData);
    }
  }, [id, baseVersion, compareVersion]);

  const left = baseData ? JSON.stringify(baseData.Definition, null, 2) : "";
  const right = compareData ? JSON.stringify(compareData.Definition, null, 2) : "";

  return (
    <div className="p-6 space-y-6">
      <h1 className="text-xl font-semibold">So sánh giữa các phiên bản</h1>

      <div className="flex gap-6">
        {/* Base Version (fixed) */}
        <div>
          <label className="block text-sm text-gray-600 mb-1">
            Phiên bản cơ sở
          </label>
          <input
            disabled
            value={baseVersion || ""}
            className="border px-3 py-2 rounded w-32 bg-gray-100"
          />
        </div>

        {/* Select version to compare */}
        <div>
          <label className="block text-sm text-gray-600 mb-1">
            So sánh với phiên bản
          </label>
          <select
            className="border px-3 py-2 rounded w-32"
            value={compareVersion || ""}
            onChange={(e) => setCompareVersion(e.target.value)}
          >
            <option value="">Chọn</option>
            {versions
              .filter(v => v !== baseVersion)
              .map(v => (
                <option key={v} value={v}>{v}</option>
              ))}
          </select>
        </div>
      </div>

      <div className="bg-white rounded shadow p-4">
        {(!baseData || !compareData) ? (
          <div className="text-gray-500">Chọn phiên bản để so sánh.</div>
        ) : (
          <DiffViewer
            oldValue={left}
            newValue={right}
            splitView={true}
            showDiffOnly={false}
          />
        )}
      </div>

      <Link
        to={`/workflows/${id}/versions`}
        className="text-primary-600 hover:text-primary-700"
      >
        ← Quay lại danh sách phiên bản
      </Link>
    </div>
  );
}
