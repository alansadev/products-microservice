package handlers

import "testing"

func TestExtractPublicIDFromURL(t *testing.T) {
	testCases := []struct {
		name     string
		inputURL string
		expected string
	}{
		{
			name:     "URL Válida",
			inputURL: "http://res.cloudinary.com/cloud-name/image/upload/v12345/sabordarondonia/arquivo123.jpg",
			expected: "sabordarondonia/arquivo123",
		},
		{
			name:     "URL com HTTPS",
			inputURL: "https://res.cloudinary.com/cloud-name/image/upload/v12345/sabordarondonia/outro_arquivo-abc.png",
			expected: "sabordarondonia/outro_arquivo-abc",
		},
		{
			name:     "URL Inválida (sem a pasta correta)",
			inputURL: "http://res.cloudinary.com/cloud-name/image/upload/v12345/outra_pasta/arquivo123.jpg",
			expected: "",
		},
		{
			name:     "URL Vazia",
			inputURL: "",
			expected: "",
		},
		{
			name:     "URL Mal Formada",
			inputURL: "isto nao e uma url",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := extractPublicIDFromURL(tc.inputURL)

			if actual != tc.expected {
				t.Errorf("actual: %s, expected: %s", actual, tc.expected)
			}
		})
	}
}
