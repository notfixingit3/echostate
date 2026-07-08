package pdf

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpubookmark "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

func countPDFPages(pdf []byte) (int, error) {
	return api.PageCount(bytes.NewReader(pdf), nil)
}

func measureCoverPages(data ReportData) (int, error) {
	m := buildReportMaroto(data.Result.Host)
	addCoverPage(m, data.Result)
	doc, err := m.Generate()
	if err != nil {
		return 0, fmt.Errorf("measure cover: %w", err)
	}
	return countPDFPages(doc.GetBytes())
}

func measureTOCStartPage(data ReportData, sections []reportSection) (int, error) {
	coverPages, err := measureCoverPages(data)
	if err != nil {
		return 0, err
	}

	m := buildReportMaroto(data.Result.Host)
	addCoverPage(m, data.Result)
	addTableOfContents(m, sections, make([]int, len(sections)))
	doc, err := m.Generate()
	if err != nil {
		return 0, fmt.Errorf("measure toc: %w", err)
	}
	combinedPages, err := countPDFPages(doc.GetBytes())
	if err != nil {
		return 0, err
	}
	if combinedPages <= coverPages {
		return coverPages, nil
	}
	return coverPages + 1, nil
}

func measureSectionPages(data ReportData, sections []reportSection) ([]int, error) {
	pages := make([]int, len(sections))
	placeholderPages := make([]int, len(sections))

	for i := range sections {
		m := buildReportMaroto(data.Result.Host)
		addCoverPage(m, data.Result)
		addTableOfContents(m, sections, placeholderPages)
		for j := 0; j < i; j++ {
			addSectionHeader(m, sections[j].title)
			sections[j].render(m)
		}
		addSectionHeader(m, sections[i].title)

		doc, err := m.Generate()
		if err != nil {
			return nil, fmt.Errorf("measure section %q: %w", sections[i].title, err)
		}

		page, err := countPDFPages(doc.GetBytes())
		if err != nil {
			return nil, fmt.Errorf("count pages for %q: %w", sections[i].title, err)
		}
		pages[i] = page
	}

	return pages, nil
}

func injectBookmarks(pdfBytes []byte, data ReportData, sections []reportSection, sectionPages []int) ([]byte, error) {
	totalPages, err := countPDFPages(pdfBytes)
	if err != nil {
		return nil, err
	}

	tocPage, err := measureTOCStartPage(data, sections)
	if err != nil {
		return nil, err
	}

	bookmarks := []pdfcpubookmark.Bookmark{
		{Title: "Cover", PageFrom: 1},
	}
	if tocPage > 0 && tocPage <= totalPages {
		bookmarks = append(bookmarks, pdfcpubookmark.Bookmark{
			Title:    "Table of Contents",
			PageFrom: tocPage,
		})
	}

	for i, section := range sections {
		if i >= len(sectionPages) {
			break
		}
		page := sectionPages[i]
		if page <= 0 || page > totalPages {
			continue
		}
		if page < bookmarks[len(bookmarks)-1].PageFrom {
			continue
		}
		bookmarks = append(bookmarks, pdfcpubookmark.Bookmark{
			Title:    section.title,
			PageFrom: page,
		})
	}

	var out bytes.Buffer
	if err := api.AddBookmarks(bytes.NewReader(pdfBytes), &out, bookmarks, true, nil); err != nil {
		return nil, fmt.Errorf("add PDF bookmarks: %w", err)
	}
	return readAllBytes(&out)
}

func readAllBytes(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
