package ste

import (
	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-azcopy/common"
	"sync/atomic"
)


// dataSchemaVersion defines the data schema version of JobPart order files supported by
// current version of azcopy
// To be Incremented every time when we release azcopy with changed dataSchema
const dataSchemaVersion = 0

type JobStatusCode uint32

// String() returns appropriate Job status in string from respective status code
func (status JobStatusCode) String() (statusString string){
	switch uint32(status){
	case 0:
		return "JobInProgress"
	case 1:
		return "JobPaused"
	case 2:
		return "JobCancelled"
	case 3:
		return "JobCompleted"
	default:
		return "InvalidStatusCode"
	}
}

const (
	// Job Part is currently executing
	JobInProgress JobStatusCode = 0

	// Job Part is currently paused and no transfer of Job is currently executing
	JobPaused JobStatusCode = 1

	// Job Part is cancelled and all transfers of the JobPart are cancelled
	JobCancelled JobStatusCode = 2

	// Job Part has completed and no transfer of JobPart is currently executing
	JobCompleted JobStatusCode = 3
)

// JobPartPlan represent the header of Job Part's Memory Map File
type JobPartPlanHeader struct {
	Version            uint32 // represent the version of data schema format of header
	Id                 common.JobID // represents the 16 byte JobId
	PartNum            uint32 // represents the part number of the JobOrder
	IsFinalPart        bool // represents whether this part is final part or not
	Priority           uint8 // represents the priority of JobPart order (High, Medium and Low)
	TTLAfterCompletion uint32 // Time to live after completion is used to persists the file on disk of specified time after the completion of JobPartOrder
	SrcLocationType    common.LocationType // represents type of source location
	DstLocationType    common.LocationType // represents type of destination location
	NumTransfers       uint32 // represents the number of transfer the JobPart order has
	LogSeverity        pipeline.LogLevel // represent the log verbosity level of logs for the specific Job
	BlobData           JobPartPlanBlobData // represent the optional attributes of JobPart Order
	// jobStatus represents the current status of JobPartPlan
	// It can have these possible values - JobInProgress, JobPaused, JobCancelled and JobCompleted
	// jobStatus is a private member whose value can be accessed by getJobStatus and setJobStatus
	jobStatus          JobStatusCode
}

// getJobStatus returns the job status stored in JobPartPlanHeader in thread-safe manner
func (jPartPlanHeader *JobPartPlanHeader) getJobStatus() (JobStatusCode){
	return JobStatusCode(atomic.LoadUint32((*uint32)(&jPartPlanHeader.jobStatus)))
}

// setJobStatus sets the job status in JobPartPlanHeader in thread-safe manner
func (jPartPlanHeader *JobPartPlanHeader)setJobStatus(status JobStatusCode) {
	atomic.StoreUint32((*uint32)(&jPartPlanHeader.jobStatus), uint32(status))
}

// JobPartPlan represent the header of Job Part's Optional Attributes in Memory Map File
type JobPartPlanBlobData struct {
	// Specifies the length of MIME content type of the blob
	ContentTypeLength     uint8
	// Specifies the MIME content type of the blob. The default type is application/octet-stream
	ContentType           [256]byte
	// Specifies length of content encoding which have been applied to the blob.
	ContentEncodingLength uint8
    // Specifies which content encodings have been applied to the blob.
	ContentEncoding       [256]byte
	MetaDataLength        uint16
	MetaData              [1000]byte
	// Specifies the maximum size of block which determines the number of chunks and chunk size of a transfer
	BlockSize             uint64
}

// JobPartPlan represent the header of Job Part's Transfer in Memory Map File
type JobPartPlanTransfer struct {
	// Offset represents the actual start offset transfer header written in JobPartOrder file
	Offset         uint64
	// SrcLength represents the actual length of source string for specific transfer
	SrcLength      uint16
	// DstLength represents the actual length of destination string for specific transfer
	DstLength      uint16
	// ChunkNum represents the num of chunks a transfer is split into
	ChunkNum       uint16
	// ModifiedTime represents the last time at which source was modified before start of transfer
	ModifiedTime   uint32
	// SourceSize represents the actual size of the source on disk
	SourceSize     uint64
	// CompletionTime represents the time at which transfer was completed
	CompletionTime uint64
	// transferStatus represents the status of current transfer (TransferInProgress, TransferFailed or TransferCompleted)
	transferStatus common.TransferStatus
}

// getTransferStatus returns the transfer status of current transfer of job part atomically
func (jPartPlanTransfer *JobPartPlanTransfer) getTransferStatus() (common.TransferStatus){
	return common.TransferStatus(atomic.LoadUint32((*uint32)(&jPartPlanTransfer.transferStatus)))
}

// getTransferStatus sets the transfer status of current transfer to given status atomically
func (jPartPlanTransfer *JobPartPlanTransfer) setTransferStatus(status common.TransferStatus){
	atomic.StoreUint32((*uint32)(&jPartPlanTransfer.transferStatus), uint32(status))
}

// These constants defines the various job priority of the JobPartOrders.
// These priorities determines the channel in which transfers will be scheduled.
const (
	HighJobPriority    = 0
	MediumJobPriority  = 1
	LowJobPriority     = 2
	DefaultJobPriority = HighJobPriority
)

const (
	MAX_SIZE_CONTENT_TYPE     = 256
	MAX_SIZE_CONTENT_ENCODING = 256
	MAX_SIZE_META_DATA        = 1000
)
