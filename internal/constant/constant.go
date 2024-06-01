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

const RawVideosDirectory = "raw-videos"
const EncodedVideosDirectory = "encoded-videos"
