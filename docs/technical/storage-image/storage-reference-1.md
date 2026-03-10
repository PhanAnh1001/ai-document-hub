# thread 1
R2 giá chắc cỡ 20$/TB nhỉ? bạn tham khảo Wasabi có 7$/TB thôi (+ mua thêm 1 CDN rẻ nữa) thì rẻ hơn đi đấy nếu lưu nhiều dung lượng.

R2 tầm 15$ bác ạ. Wasabi mình thấy giá cũng ổn, băng thông không giới hạn, hạn chế nhất lại ko có CDN, sản phẩm mình hơi đặc thù, chi phí cộng dồn mua cdn có khi lại cao hơn thằng R2.

mình dùng Bunny CDN, giá tầm $0.03/GB --> nếu bandwidth 100Gb/tháng thì hết có 3$.

# thread 2
Discord CDN free ko giới hạn gì.
Vấn đề là mỗi lần lấy phải refresh lại link tệp tin. Nếu số lượng cần lấy cùng lúc lớn thì thời gian chờ lâu quá.

# thread 3
Storage cost optimization là bài toán quan trọng với startup. Với sản phẩm lưu trữ ảnh, bạn nên phân tích: 1) Tỷ lệ access frequency (hot/cold data) để chọn tier phù hợp, 2) Implement CDN caching hiệu quả, 3) Sử dụng compression và format tối ưu như WebP/AVIF. R2 tốt nhưng nếu storage chiếm >20% chi phí, cân nhắc hybrid solution: hot data trên R2, cold data trên B2/Wasabi. Ở VN có thể tham khảo VNG Cloud Storage hoặc Viettel IDC nhưng cần test SLA kỹ.
