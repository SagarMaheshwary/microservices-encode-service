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
	S3ThumbnailsDirectory    = "thumbnails"
)

const TempVideosDownloadDirectory = "videos"

const MPEGDASHManifestFile = "master.mpd"

const ServiceName = "Encode Service"

const (
	TraceTypeRabbitMQConsume = "RabbitMQ Consume"
	TraceTypeRabbitMQPublish = "RabbitMQ Publish"
)
