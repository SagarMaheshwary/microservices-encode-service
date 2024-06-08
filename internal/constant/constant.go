package constant

const (
	QueueEncodeService       = "EncodeService"
	QueueVideoCatalogService = "VideoCatalogService"
)

const (
	MessageTypeEncodeUploadedVideo    = "EncodeUploadedVideo"
	MessageTypeVideoEncodingCompleted = "VideoEncodingCompleted"
)

const (
	ContentTypeJSON = "application/json"
)

const (
	ProtocolTCP  = "tcp"
	ProtocolAMQP = "amqp"
)

const (
	S3RawVideosDirectory     = "raw-videos"
	S3EncodedVideosDirectory = "encoded-videos"
)

const (
	ExtensionMPEGDASH = "mpd"
)

const TempVideosDownloadDirectory = "assets/videos"
