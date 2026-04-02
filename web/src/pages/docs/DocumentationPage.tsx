import React, { type ReactNode } from 'react';
import { Lightbulb } from 'lucide-react';

const CodeBlock = ({ children }: { children: ReactNode }) => (
  <pre className="bg-gray-100 p-3 rounded-md text-sm text-gray-800 overflow-x-auto">
    <code>{children}</code>
  </pre>
);

const SectionTitle = ({ children }: { children: ReactNode }) => (
  <h2 className="text-2xl font-bold text-gray-800 mt-8 mb-4 border-b pb-2">{children}</h2>
);

const SubTitle = ({ children }: { children: ReactNode }) => (
  <h3 className="text-xl font-semibold text-gray-700 mt-6 mb-3">{children}</h3>
);

const Note = ({ children }: { children: ReactNode }) => (
  <div className="flex items-start space-x-3 bg-yellow-50 border border-yellow-200 text-yellow-800 p-4 rounded-lg my-4">
    <Lightbulb className="w-5 h-5 mt-1 flex-shrink-0" />
    <div>{children}</div>
  </div>
);

export default function DocumentationPage() {
  return (
    <div className="p-6 bg-white max-w-4xl mx-auto rounded-lg shadow">
      <h1 className="text-4xl font-bold text-primary-600 mb-4">Hướng dẫn sử dụng</h1>
      <p className="text-lg text-gray-600">
        Chào mừng đến với tài liệu hướng dẫn sử dụng hệ thống tự động hóa Tempo. Dưới đây là hướng dẫn chi tiết về cách cấu hình và sử dụng từng loại node trong workflow của bạn.
      </p>

      <SectionTitle>Biến Template</SectionTitle>
      <p>Hệ thống hỗ trợ sử dụng biến template để truyền dữ liệu giữa các action. Bạn có thể truy cập kết quả của các node trước đó bằng cách sử dụng ID của chúng.</p>
      <p className="mt-2">Cú pháp chung: {"`{{ .node_id.path.to.value }}`"}</p>
      <CodeBlock>
        {`// Lấy toàn bộ body của trigger
{{ .trigger.data }}

// Lấy trường 'name' từ body của trigger
{{ .trigger.data.name }}

// Lấy trường 'file_path' từ kết quả của action có ID là 'create_excel'
{{ .create_excel.file_path }}`}
      </CodeBlock>
      <Note>Để sử dụng dữ liệu là một mảng hoặc object (ví dụ: cho action Excel), bạn cần sử dụng pipe `| json` để đảm bảo định dạng đúng: {"`{{ .http_action.body | json }}`"}</Note>

      <SectionTitle>Triggers (Bộ kích hoạt)</SectionTitle>
      
      <SubTitle>Webhook</SubTitle>
      <p>Kích hoạt workflow khi có một yêu cầu HTTP POST được gửi đến một URL duy nhất.</p>
      <ul>
        <li><strong>Webhook Token (tùy chọn):</strong> Một chuỗi bí mật để xác thực yêu cầu. Nếu được đặt, URL sẽ cần có tham số truy vấn `?token=your_secret_token`.</li>
      </ul>

      <SubTitle>Cron (Lên lịch)</SubTitle>
      <p>Kích hoạt workflow theo một lịch trình cố định.</p>
      <ul>
        <li><strong>Expression:</strong> Chuỗi cron expression theo chuẩn 5 thành phần: `phút giờ ngày-trong-tháng tháng ngày-trong-tuần`.</li>
      </ul>
      <p className="mt-2">Ví dụ:</p>
      <CodeBlock>
        {`# Chạy vào 9 giờ sáng mỗi ngày
0 9 * * *

# Chạy mỗi 15 phút
*/15 * * * *`}
      </CodeBlock>

      <SectionTitle>Actions (Hành động)</SectionTitle>

      <SubTitle>HTTP</SubTitle>
      <p>Gửi một yêu cầu HTTP đến một URL bất kỳ.</p>
      <ul>
        <li><strong>Phương thức:</strong> GET, POST, PUT, PATCH, DELETE.</li>
        <li><strong>Đường dẫn URL:</strong> URL của API bạn muốn gọi. Có thể chứa biến template.</li>
        <li><strong>Headers (JSON):</strong> Cấu trúc JSON cho các header của request.</li>
        <li><strong>Body (JSON):</strong> Cấu trúc JSON cho body của request.</li>
      </ul>

      <SubTitle>Email</SubTitle>
      <p>Gửi một email thông qua SMTP đã được cấu hình.</p>
      <ul>
        <li><strong>Tới Email:</strong> Địa chỉ email người nhận. Ví dụ: {"`{{ .trigger.data.email }}`"}.</li>
        <li><strong>Tiêu đề:</strong> Tiêu đề của email.</li>
        <li><strong>Nội dung:</strong> Nội dung của email.</li>
      </ul>

      <SubTitle>Excel</SubTitle>
      <p>Tạo một file `.xlsx` từ một mảng dữ liệu.</p>
      <ul>
        <li><strong>Tên file:</strong> Tên file Excel sẽ được tạo. Ví dụ: {"`report-{{ .trigger.data.date }}.xlsx`"}.</li>
        <li><strong>Dữ liệu:</strong> Một biến template trỏ đến một **mảng các object**. Ví dụ: {"`{{ .http_get_users.body | json }}`"}.</li>
        <li><strong>Tiêu đề cột:</strong> Danh sách các cột, cách nhau bởi dấu phẩy. Ví dụ: `id,name,email`.</li>
      </ul>

      <SubTitle>Google Drive</SubTitle>
      <p>Tải một file hoặc nội dung lên Google Drive.</p>
      <ul>
        <li><strong>Google Integration ID:</strong> ID của tích hợp Google bạn đã kết nối trong trang "Tích hợp".</li>
        <li><strong>Filename:</strong> Tên file sẽ được lưu trên Google Drive.</li>
        <li><strong>Parent Folder ID (tùy chọn):</strong> ID của thư mục trên Drive để tải file vào.</li>
        <li><strong>Content Type:</strong> MIME type của file, ví dụ: `text/plain`, `application/pdf`.</li>
        <li><strong>File Path (tùy chọn):</strong> Đường dẫn đến một file đã được tạo bởi action trước (ví dụ: action Excel). Ví dụ: {"`{{ .create_excel_report.file_path }}`"}.</li>
        <li><strong>Content (nếu không dùng File Path):</strong> Nội dung văn bản để tạo file mới.</li>
      </ul>

      <SubTitle>Google Sheets</SubTitle>
      <p>Thêm một hàng mới vào một trang tính Google.</p>
      <ul>
        <li><strong>Google Integration ID:</strong> ID của tích hợp Google.</li>
        <li><strong>Spreadsheet ID:</strong> ID của file Google Sheet (lấy từ URL).</li>
        <li><strong>Sheet Name:</strong> Tên của trang tính (tab) cần ghi dữ liệu.</li>
        <li><strong>Dữ liệu hàng (Row Data):</strong> Một biến template trỏ đến một **mảng các giá trị**. Ví dụ: {"`{{ [ .trigger.data.name, .trigger.data.email ] | json }}`"}.</li>
      </ul>

      <SubTitle>Google Cloud Storage (GCS)</SubTitle>
      <p>Tải file hoặc nội dung lên một bucket trên GCS.</p>
      <ul>
        <li><strong>Google Integration ID:</strong> ID của tích hợp Google.</li>
        <li><strong>Bucket Name:</strong> Tên của GCS bucket.</li>
        <li><strong>Object Name:</strong> Tên/đường dẫn của file trong bucket.</li>
        <li><strong>File Path / Content:</strong> Tương tự như Google Drive.</li>
      </ul>

      <SubTitle>Notion</SubTitle>
      <p>Tạo một trang mới trong một database trên Notion.</p>
      <ul>
        <li><strong>Notion Integration ID:</strong> ID của tích hợp Notion.</li>
        <li><strong>Database ID:</strong> ID của database trên Notion.</li>
        <li><strong>Properties (JSON):</strong> Cấu trúc JSON phức tạp để định nghĩa các thuộc tính của trang. Cần tuân thủ đúng định dạng của Notion API.</li>
        <li><strong>Content (Markdown - tùy chọn):</strong> Nội dung cho trang, viết dưới dạng Markdown đơn giản.</li>
      </ul>
      
      <SubTitle>GitHub & Discord</SubTitle>
      <p>Các action này cho phép bạn tương tác với GitHub (tạo issue, PR) và Discord (gửi tin nhắn, embed). Cấu hình của chúng khá trực quan trong modal, bao gồm việc chọn `action` cụ thể và điền các tham số tương ứng.</p>

    </div>
  );
}