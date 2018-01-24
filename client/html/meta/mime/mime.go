package mime

type Mime struct {
	value string
}

var (
	TextHtml = Mime{"text/html"}
	TextCss = Mime{"text/css"}
	TextCsv = Mime{"text/csv"}
	Javascript = Mime{"application/javascript"}
	Multipart = Mime{"multipart/form-data"}
	MessagePartial = Mime{"message/partial"}
	ImagePng = Mime{"image/png"}
	ImageGif = Mime{"image/gif"}
	ImageJpeg = Mime{"image/jpeg"}
	ImageIcon = Mime{"image/x-icon"}
	ImageTiff = Mime{"image/tiff"}
	AudioMpeg = Mime{"audio/mpeg"}
	VideoMp4 = Mime{"video/mp4"}
	ArchiveRar = Mime{"application/x-rar-compressed"}
	ArchiveTar = Mime{"application/x-tar"}
	ArchiveZip = Mime{"application/zip"}
	Archive7z = Mime{"application/x-7z-compressed"}
	ArchiveBzip = Mime{"application/x-bzip"}
	ArchiveBzip2 = Mime{"application/x-bzip2"}
	FormatJson = Mime{"application/json"}
	FormatXml = Mime{"application/xml"}
	FormatBinary = Mime{"application/octet-stream"}
	FormatPdf = Mime{"application/pdf"}
	FormatShell = Mime{"application/x-sh"}
	FontTtf = Mime{"font/ttf"}
	FontWoff = Mime{"font/woff"}
	FontWoff2 = Mime{"font/woff2"}
)

func (contentType Mime) String() string {
	return contentType.value
}