package contentextractor

import (
	"context"
	"testing"
)

func TestResourceProcessor_pdfToMD(t *testing.T) {
	// Минимальный валидный PDF (одна пустая страница)
	pdfData := []byte(`%PDF-1.4
%âãÏÓ
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj
xref
0 4
0000000000 65535 f 
0000000010 00000 n 
0000000053 00000 n 
0000000104 00000 n 
trailer
<< /Size 4 /Root 1 0 R >>
startxref
178
%%EOF`)

	ctx := context.Background()
	processor := &ContentExtractor{}

	md, err := processor.pdfToMD(ctx, pdfData)
	if err != nil {
		t.Fatalf("pdfToMD вернула ошибку: %v", err)
	}

	if len(md) == 0 {
		t.Errorf("pdfToMD вернула пустой результат")
	}
}
