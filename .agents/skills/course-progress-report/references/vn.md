# Hướng dẫn viết báo cáo tiếng Việt

Viết bản ngắn gọn cho giảng viên từ detail report đã có. Dùng bản detail và các chỉ dẫn bổ sung của người dùng làm nguồn; không quét lại Git, mở các liên kết nguồn hay đọc lại nhật ký UAT. Nếu chưa có bản detail tương ứng, hỏi người dùng thay vì tự viết cả hai bản.

## 1. Mở đầu

- Lưu tại `docs/src/reports/YYYY-MM-dd.md`. Giữ `authorship` là `ai` trong frontmatter và dùng `# YYYY-MM-dd` làm tiêu đề cấp 1. Không thêm `title` hoặc lặp lại tiêu đề dưới dạng chữ đậm.
- Ghi **Kỳ báo cáo**, **Commit mốc (không bao gồm)** và **Commit cuối trên nhánh tương ứng (bao gồm)**. Dùng cùng phạm vi và commit với bản detail; gắn URL đầy đủ cho cả hai commit.
- Viết một đoạn tóm tắt ngắn về kết quả chính và tình trạng hiện tại. Giữ phạm vi loại trừ nếu cần để tránh lặp lại nội dung đã báo cáo.

## 2. Nội dung

Đánh số mọi đề mục bên dưới H1 theo cấp, ví dụ `## 1. Kết quả thực hiện` và `### 1.1. Thanh toán`. Linh hoạt chọn mục theo nội dung từng kỳ. Có thể dùng bố cục sau:

1. **Kết quả thực hiện:** Tách các đầu việc có kết quả hoặc nguồn đối chiếu riêng. Mỗi ý nói ngắn gọn đã làm gì và vì sao đáng chú ý. Giữ chi tiết kỹ thuật khi nó giúp giảng viên hiểu kết quả; không dịch nguyên bản detail.
2. **Kiểm thử và việc đang thực hiện:** Tóm tắt UAT trong vài câu, nêu kết quả chính và phần chưa kiểm thử đầy đủ. Giữ liên kết tới kịch bản và nhật ký nếu có. Phân biệt đã chuẩn bị triển khai với đã triển khai và xác minh.
3. **Kế hoạch:** Tách **giai đoạn tiếp theo** (công việc trước mắt) và **học kỳ tới** (định hướng dài hạn). Phân biệt đề xuất, việc chờ điều kiện và việc đã làm.

## 3. Thuật ngữ kỹ thuật

Viết câu giải thích bằng tiếng Việt, nhưng giữ nguyên thuật ngữ kỹ thuật tiếng Anh quen thuộc trong dự án. Dùng nhất quán trong cả tiêu đề và nội dung; nếu cần, giải thích ngắn bằng tiếng Việt thay vì đặt một tên dịch mới.

| Giữ nguyên    | Thay vì dịch thành |
| ------------- | ------------------ |
| contract API  | hợp đồng API       |
| event         | sự kiện            |
| config        | cấu hình           |
| collaboration | cộng tác           |
| publish       | xuất bản           |
| workspace     | không gian lưu trữ |

## 4. Tham chiếu và kiểm tra

- Đặt nguồn cạnh đầu việc tương ứng. Dùng nhãn như `[notopia-uit/notopia#365](https://github.com/notopia-uit/notopia/pull/365)` cho PR. Hash có thể hiện ngắn trong nhãn nhưng phải được gắn URL commit đầy đủ. Tạo URL từ repo và hash đã ghi trong bản detail, không cần mở link.
- Giữ các link demo, Gist, issue và tài liệu mà người dùng muốn nhắc tới. Có thể để URL trần nếu không có nhãn phù hợp. Nêu rõ demo chỉ để tham khảo nếu chưa tích hợp.
- Cập nhật `docs/src/reports/index.md`, giữ nguyên bản detail và các chỉnh sửa của người dùng. Kiểm tra tên file, tiêu đề, số đề mục và liên kết nội bộ; không chạy build docs.
