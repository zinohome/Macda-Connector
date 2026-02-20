package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
)

var beijingLoc *time.Location

func init() {
	var err error
	beijingLoc, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// fallback: +8
		beijingLoc = time.FixedZone("CST", 8*3600)
	}
}

type NB67Processor struct {
	count          int64
	logSampleEvery int64
}

type ParsedOutput struct {
	HeaderCode01    uint8  `json:"header_code_01"`
	HeaderCode02    uint8  `json:"header_code_02"`
	MessageLength   uint16 `json:"message_length"`
	SrcDeviceNo     uint8  `json:"src_device_no"`
	HostDeviceNo    uint8  `json:"host_device_no"`
	MessageType     uint16 `json:"message_type"`
	FrameNo         uint16 `json:"frame_no"`
	LineNo          uint16 `json:"line_no"`
	TrainType       uint16 `json:"train_type"`
	TrainNo         uint32 `json:"train_no"`
	CarriageNo      uint8  `json:"carriage_no"`
	ProtocolVersion uint8  `json:"protocol_version"`

	SrcYear   uint8 `json:"src_year"`
	SrcMonth  uint8 `json:"src_month"`
	SrcDay    uint8 `json:"src_day"`
	SrcHour   uint8 `json:"src_hour"`
	SrcMinute uint8 `json:"src_minute"`
	SrcSecond uint8 `json:"src_second"`

	DvcFlag     uint8  `json:"dvc_flag"`
	DvcTrainNo  uint16 `json:"dvc_train_no"`
	DvcCarriage uint8  `json:"dvc_carriage_no"`
	DvcYear     uint8  `json:"dvc_year"`
	DvcMonth    uint8  `json:"dvc_month"`
	DvcDay      uint8  `json:"dvc_day"`
	DvcHour     uint8  `json:"dvc_hour"`
	DvcMinute   uint8  `json:"dvc_minute"`
	DvcSecond   uint8  `json:"dvc_second"`

	StatusVentilationU1 bool `json:"status_ventilation_u1"`
	StatusCoolingU1     bool `json:"status_cooling_u1"`
	StatusCompressorU11 bool `json:"status_compressor_u11"`
	StatusCompressorU12 bool `json:"status_compressor_u12"`
	StatusAirPurifierU1 bool `json:"status_air_purifier_u1"`

	Tveh1        int16 `json:"tveh_1"`
	Humdity1     int16 `json:"humdity_1"`
	Tveh2        int16 `json:"tveh_2"`
	Humdity2     int16 `json:"humdity_2"`
	AqTU1        int16 `json:"aq_t_u1"`
	AqHU1        int16 `json:"aq_h_u1"`
	AqCo2U1      int16 `json:"aq_co2_u1"`
	AqTvocU1     int16 `json:"aq_tvoc_u1"`
	AqPm25U1     int16 `json:"aq_pm2_5_u1"`
	AqPm10U1     int16 `json:"aq_pm10_u1"`
	FCpU11       int16 `json:"f_cp_u11"`
	ICpU11       int16 `json:"i_cp_u11"`
	VCpU11       int16 `json:"v_cp_u11"`
	PCpU11       int16 `json:"p_cp_u11"`
	SucktU11     int16 `json:"suckt_u11"`
	HighpressU11 int16 `json:"highpress_u11"`

	BlpfltCompU11 bool `json:"blpflt_comp_u11"`
	BscfltCompU11 bool `json:"bscflt_comp_u11"`
	BscfltVentU11 bool `json:"bscflt_vent_u11"`
	BfltFadU11    bool `json:"bflt_fad_u11"`
	BfltRadU11    bool `json:"bflt_rad_u11"`

	DmpExuPos       uint16 `json:"dmp_exu_pos"`
	StartStation    uint16 `json:"start_station"`
	TerminalStation uint16 `json:"terminal_station"`
	CurStation      uint16 `json:"cur_station"`
	NextStation     uint16 `json:"next_station"`

	ParserVersion  string `json:"parser_version"`
	QualityStatus  string `json:"quality_status"`
	FrameSize      int    `json:"frame_size"`
	ParsedAtUnixMs int64  `json:"parsed_at_unix_ms"`

	ParsedAt string `json:"parsed_at"`

	// Raw contains the full Kaitai-decoded struct for completeness.
	Raw *Nb67 `json:"raw,omitempty"`
}

func NewNB67Processor(conf *service.ParsedConfig) (*NB67Processor, error) {
	logSampleEvery := int64(100)
	if v, err := conf.FieldInt("log_sample_every"); err == nil {
		logSampleEvery = int64(v)
	}
	return &NB67Processor{logSampleEvery: logSampleEvery}, nil
}

func (p *NB67Processor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error) {
	payload, err := msg.AsBytes()
	if err != nil {
		return service.MessageBatch{msg}, fmt.Errorf("failed to get message bytes: %w", err)
	}

	nb67 := &Nb67{}
	io := kaitai.NewStream(bytes.NewReader(payload))
	if err := nb67.Read(io, nil, nb67); err != nil {
		return service.MessageBatch{msg}, fmt.Errorf("NB67 parse error: %w", err)
	}

	now := time.Now().In(beijingLoc)
	output := &ParsedOutput{
		HeaderCode01:    nb67.MsgHeaderCode01,
		HeaderCode02:    nb67.MsgHeaderCode02,
		MessageLength:   nb67.MsgLength,
		SrcDeviceNo:     nb67.MsgSrcDvcNo,
		HostDeviceNo:    nb67.MsgHostDvcNo,
		MessageType:     nb67.MsgType,
		FrameNo:         nb67.MsgFrameNo,
		LineNo:          nb67.MsgLineNo,
		TrainType:       nb67.MsgTrainType,
		TrainNo:         nb67.MsgTrainNo,
		CarriageNo:      nb67.MsgCarriageNo,
		ProtocolVersion: nb67.MsgProtocalVersion,

		SrcYear:   nb67.MsgSrcDvcYear,
		SrcMonth:  nb67.MsgSrcDvcMonth,
		SrcDay:    nb67.MsgSrcDvcDay,
		SrcHour:   nb67.MsgSrcDvcHour,
		SrcMinute: nb67.MsgSrcDvcMinute,
		SrcSecond: nb67.MsgSrcDvcSecond,

		DvcFlag:     nb67.DvcFlag,
		DvcTrainNo:  nb67.DvcTrainNo,
		DvcCarriage: nb67.DvcCarriageNo,
		DvcYear:     nb67.DvcYear,
		DvcMonth:    nb67.DvcMonth,
		DvcDay:      nb67.DvcDay,
		DvcHour:     nb67.DvcHour,
		DvcMinute:   nb67.DvcMinute,
		DvcSecond:   nb67.DvcSecond,

		StatusVentilationU1: nb67.CfbkEfU11,
		StatusCoolingU1:     nb67.CfbkCfU11,
		StatusCompressorU11: nb67.CfbkCompU11,
		StatusCompressorU12: nb67.CfbkCompU12,
		StatusAirPurifierU1: nb67.CfbkApU11,

		Tveh1:        nb67.Tveh1,
		Humdity1:     nb67.Humdity1,
		Tveh2:        nb67.Tveh2,
		Humdity2:     nb67.Humdity2,
		AqTU1:        nb67.AqTU1,
		AqHU1:        nb67.AqHU1,
		AqCo2U1:      nb67.AqCo2U1,
		AqTvocU1:     nb67.AqTvocU1,
		AqPm25U1:     nb67.AqPm25U1,
		AqPm10U1:     nb67.AqPm10U1,
		FCpU11:       nb67.FCpU11,
		ICpU11:       nb67.ICpU11,
		VCpU11:       nb67.VCpU11,
		PCpU11:       nb67.PCpU11,
		SucktU11:     nb67.SucktU11,
		HighpressU11: nb67.HighpressU11,

		BlpfltCompU11: nb67.BlpfltCompU11,
		BscfltCompU11: nb67.BscfltCompU11,
		BscfltVentU11: nb67.BscfltVentU11,
		BfltFadU11:    nb67.BfltFadU11,
		BfltRadU11:    nb67.BfltRadU11,

		DmpExuPos:       nb67.DmpExuPos,
		StartStation:    nb67.StartStation,
		TerminalStation: nb67.TerminalStation,
		CurStation:      nb67.CurStation,
		NextStation:     nb67.NextStation,

		ParserVersion:  "nb67-v1",
		QualityStatus:  "OK",
		FrameSize:      len(payload),
		ParsedAtUnixMs: now.UnixMilli(),
		ParsedAt:       now.Format(time.RFC3339Nano),
		Raw:            nb67,
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		return service.MessageBatch{msg}, fmt.Errorf("JSON marshal error: %w", err)
	}

	msg.SetBytes(jsonBytes)

	p.count++
	if p.logSampleEvery > 0 && p.count%p.logSampleEvery == 0 {
		log.Printf("[NB67] Processed %d frames: TrainNo=%d Carriage=%d CurStation=%d", p.count, output.TrainNo, output.CarriageNo, output.CurStation)
	}

	return service.MessageBatch{msg}, nil
}

func (p *NB67Processor) Close(ctx context.Context) error {
	return nil
}
