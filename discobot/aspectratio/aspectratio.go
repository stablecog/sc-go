package aspectratio

import (
	"errors"

	"github.com/google/uuid"
)

// Hacky hardcoded stuff but it gets the job done
const LUNA_ID = "b6c1372f-31a7-457c-907c-d292a6ffef97"
const _22_DIFFUSION_ID = "fc06f6ab-ed14-4186-a7c0-aaec288d4f38"
const WAIFU_ID = "f7f3d973-ac6f-4a7a-9db8-e89e4fba03a9"

var AvailableRatios = []AspectRatio{
	AspectRatio_1_1,
	AspectRatio_2_3,
	AspectRatio_3_2,
	AspectRatio_9_16,
	AspectRatio_16_9,
	AspectRatio_4_5,
	AspectRatio_2dot4_1,
}

var DefaultAspectRatio = AspectRatio_2_3

type AspectRatio int

const (
	AspectRatio_16_9 AspectRatio = iota
	AspectRatio_1_1
	AspectRatio_2_3
	AspectRatio_3_2
	AspectRatio_9_16
	AspectRatio_4_5
	AspectRatio_2dot4_1
)

func (a AspectRatio) String() string {
	var ratio string
	switch a {
	case AspectRatio_16_9:
		ratio = "Desktop (16:9)"
	case AspectRatio_1_1:
		ratio = "Square (1:1)"
	case AspectRatio_2_3:
		ratio = "Portrait (2:3)"
	case AspectRatio_3_2:
		ratio = "Landscape (3:2)"
	case AspectRatio_9_16:
		ratio = "Mobile (9:16)"
	case AspectRatio_4_5:
		ratio = "Squarish (4:5)"
	case AspectRatio_2dot4_1:
		ratio = "Anamorphic (2.4:1)"
	default:
		ratio = "Unknown"
	}
	if a == DefaultAspectRatio {
		return ratio + " (default)"
	}
	return ratio
}

func (a AspectRatio) SimpleString() string {
	switch a {
	case AspectRatio_16_9:
		return "16-9"
	case AspectRatio_1_1:
		return "1-1"
	case AspectRatio_2_3:
		return "2-3"
	case AspectRatio_3_2:
		return "3-2"
	case AspectRatio_9_16:
		return "9-16"
	case AspectRatio_4_5:
		return "4-5"
	case AspectRatio_2dot4_1:
		return "2_4-1"
	default:
		return "unknown"
	}
}

func GetAspectRatioBySimpleString(simpleString string) (AspectRatio, error) {
	switch simpleString {
	case "16-9":
		return AspectRatio_16_9, nil
	case "1-1":
		return AspectRatio_1_1, nil
	case "2-3":
		return AspectRatio_2_3, nil
	case "3-2":
		return AspectRatio_3_2, nil
	case "9-16":
		return AspectRatio_9_16, nil
	case "4-5":
		return AspectRatio_4_5, nil
	case "2_4-1":
		return AspectRatio_2dot4_1, nil
	default:
		return DefaultAspectRatio, errors.New("unknown aspect ratio")
	}
}

func (a AspectRatio) GetAllWidthHeightCombos() (widths, heights []int32) {
	switch a {
	case AspectRatio_16_9:
		return []int32{1280, 1024, 768}, []int32{720, 576, 432}
	case AspectRatio_1_1:
		return []int32{1024, 768, 512}, []int32{1024, 768, 512}
	case AspectRatio_2_3:
		return []int32{832, 608, 512}, []int32{1248, 912, 768}
	case AspectRatio_3_2:
		return []int32{1248, 912, 768}, []int32{832, 608, 512}
	case AspectRatio_9_16:
		return []int32{720, 576, 432}, []int32{1280, 1024, 768}
	case AspectRatio_4_5:
		return []int32{896, 672, 512}, []int32{1120, 840, 640}
	case AspectRatio_2dot4_1:
		return []int32{1536, 1152, 768}, []int32{640, 480, 320}
	default:
		return
	}
}

func (a AspectRatio) GetWidthHeightForModel(modelId uuid.UUID) (width, height int32) {
	switch a {
	case AspectRatio_16_9:
		switch modelId.String() {
		case WAIFU_ID, _22_DIFFUSION_ID:
			return 768, 432
		case LUNA_ID:
			return 1024, 576
		default:
			return 1280, 720
		}
	case AspectRatio_1_1:
		switch modelId.String() {
		case WAIFU_ID, _22_DIFFUSION_ID:
			return 512, 512
		case LUNA_ID:
			return 768, 768
		default:
			return 1024, 1024
		}
	case AspectRatio_2_3:
		switch modelId.String() {
		case WAIFU_ID, _22_DIFFUSION_ID:
			return 512, 768
		case LUNA_ID:
			return 608, 912
		default:
			return 832, 1248
		}
	case AspectRatio_3_2:
		switch modelId.String() {
		case WAIFU_ID, _22_DIFFUSION_ID:
			return 768, 512
		case LUNA_ID:
			return 912, 608
		default:
			return 1248, 832
		}
	case AspectRatio_9_16:
		switch modelId.String() {
		case WAIFU_ID, _22_DIFFUSION_ID:
			return 432, 768
		case LUNA_ID:
			return 576, 1024
		default:
			return 720, 1280
		}
	case AspectRatio_4_5:
		switch modelId.String() {
		case WAIFU_ID, _22_DIFFUSION_ID:
			return 512, 640
		case LUNA_ID:
			return 672, 840
		default:
			return 896, 1120
		}
	case AspectRatio_2dot4_1:
		switch modelId.String() {
		case WAIFU_ID, _22_DIFFUSION_ID:
			return 768, 320
		case LUNA_ID:
			return 1152, 480
		default:
			return 1536, 640
		}
	default:
		return
	}
}
