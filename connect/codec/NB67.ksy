meta:
  id: nb67
  endian: be
seq:
  - id: msg_header_code01
    type: u1
  - id: msg_header_code02
    type: u1
  - id: msg_length
    type: u2
  - id: msg_src_dvc_no
    type: u1
  - id: msg_host_dvc_no
    type: u1
  - id: msg_type
    type: u2
  - id: msg_frame_no
    type: u2
  - id: msg_line_no
    type: u2
  - id: msg_train_type
    type: u2
  - id: msg_train_no
    type: u4
  - id: msg_carriage_no
    type: u1
  - id: msg_protocal_version
    type: u1
  - id: msg_reversed1
    type: u2
  - id: msg_reversed2
    type: u2
  - id: msg_reversed3
    type: u2
  - id: msg_reversed4
    type: u2
  - id: msg_reversed5
    type: u2
  - id: msg_src_dvc_year
    type: u1
  - id: msg_src_dvc_month
    type: u1
  - id: msg_src_dvc_day
    type: u1
  - id: msg_src_dvc_hour
    type: u1
  - id: msg_src_dvc_minute
    type: u1
  - id: msg_src_dvc_second
    type: u1
  - id: dvc_flag
    type: u1
  - id: dvc_train_no
    type: u2
  - id: dvc_carriage_no
    type: u1
  - id: dvc_year
    type: u1
  - id: dvc_month
    type: u1
  - id: dvc_day
    type: u1
  - id: dvc_hour
    type: u1
  - id: dvc_minute
    type: u1
  - id: dvc_second
    type: u1
  - id: ig_rsv0
    type: u1
  - id: ig_rsv1
    type: u1
  - id: cfbk_ef_u11
    type: b1le
  - id: ig_rsv2
    type: b1le
  - id: cfbk_cf_u11
    type: b1le
  - id: ig_rsv3
    type: b1le
  - id: cfbk_comp_u11
    type: b1le
  - id: cfbk_comp_u12
    type: b1le
  - id: cfbk_ap_u11
    type: b1le
  - id: ig_rsv4
    type: b1le
  - id: cfbk_ef_u21
    type: b1le
  - id: ig_rsv5
    type: b1le
  - id: cfbk_cf_u21
    type: b1le
  - id: ig_rsv6
    type: b1le
  - id: cfbk_comp_u21
    type: b1le
  - id: cfbk_comp_u22
    type: b1le
  - id: cfbk_ap_u21
    type: b1le
  - id: ig_rsv7
    type: b1le
  - id: cfbk_tpp_u1
    type: b1le
  - id: cfbk_tpp_u2
    type: b1le
  - id: cfbk_ev_u1
    type: b1le
  - id: cfbk_ev_u2
    type: b1le
  - id: cfbk_ewd
    type: b1le
  - id: cfbk_exufan
    type: b1le
  - id: ig_rsv9
    type: b1le
  - id: ig_rsv10
    type: b1le
  - id: bocflt_ef_u11
    type: b1le
  - id: bocflt_ef_u12
    type: b1le
  - id: bocflt_cf_u11
    type: b1le
  - id: bocflt_cf_u12
    type: b1le
  - id: bflt_vfd_u11
    type: b1le
  - id: bflt_vfd_com_u11
    type: b1le
  - id: bflt_vfd_u12
    type: b1le
  - id: bflt_vfd_com_u12
    type: b1le
  - id: blpflt_comp_u11
    type: b1le
  - id: bscflt_comp_u11
    type: b1le
  - id: bscflt_vent_u11
    type: b1le
  - id: blpflt_comp_u12
    type: b1le
  - id: bscflt_comp_u12
    type: b1le
  - id: bscflt_vent_u12
    type: b1le
  - id: bflt_fad_u11
    type: b1le
  - id: bflt_fad_u12
    type: b1le
  - id: ig_rsv11
    type: b1le
  - id: ig_rsv12
    type: b1le
  - id: bflt_rad_u11
    type: b1le
  - id: bflt_rad_u12
    type: b1le
  - id: ig_rsv13
    type: b1le
  - id: ig_rsv14
    type: b1le
  - id: bflt_ap_u11
    type: b1le
  - id: ig_rsv15
    type: b1le
  - id: bflt_expboard_u1
    type: b1le
  - id: bflt_frstemp_u1
    type: b1le
  - id: bflt_rnttemp_u1
    type: b1le
  - id: bflt_splytemp_u11
    type: b1le
  - id: bflt_splytemp_u12
    type: b1le
  - id: bflt_coiltemp_u11
    type: b1le
  - id: bflt_coiltemp_u12
    type: b1le
  - id: bflt_insptemp_u11
    type: b1le
  - id: bflt_insptemp_u12
    type: b1le
  - id: bflt_lowpres_u11
    type: b1le
  - id: bflt_lowpres_u12
    type: b1le
  - id: bflt_highpres_u11
    type: b1le
  - id: bflt_highpres_u12
    type: b1le
  - id: bflt_diffpres_u1
    type: b1le
  - id: bocflt_ef_u21
    type: b1le
  - id: bocflt_ef_u22
    type: b1le
  - id: bocflt_cf_u21
    type: b1le
  - id: bocflt_cf_u22
    type: b1le
  - id: bflt_vfd_u21
    type: b1le
  - id: bflt_vfd_com_u21
    type: b1le
  - id: bflt_vfd_u22
    type: b1le
  - id: bflt_vfd_com_u22
    type: b1le
  - id: blpflt_comp_u21
    type: b1le
  - id: bscflt_comp_u21
    type: b1le
  - id: bscflt_vent_u21
    type: b1le
  - id: blpflt_comp_u22
    type: b1le
  - id: bscflt_comp_u22
    type: b1le
  - id: bscflt_vent_u22
    type: b1le
  - id: bflt_fad_u21
    type: b1le
  - id: bflt_fad_u22
    type: b1le
  - id: ig_rsv16
    type: b1le
  - id: ig_rsv17
    type: b1le
  - id: bflt_rad_u21
    type: b1le
  - id: bflt_rad_u22
    type: b1le
  - id: ig_rsv18
    type: b1le
  - id: ig_rsv19
    type: b1le
  - id: bflt_ap_u21
    type: b1le
  - id: ig_rsv20
    type: b1le
  - id: bflt_expboard_u2
    type: b1le
  - id: bflt_frstemp_u2
    type: b1le
  - id: bflt_rnttemp_u2
    type: b1le
  - id: bflt_splytemp_u21
    type: b1le
  - id: bflt_splytemp_u22
    type: b1le
  - id: bflt_coiltemp_u21
    type: b1le
  - id: bflt_coiltemp_u22
    type: b1le
  - id: bflt_insptemp_u21
    type: b1le
  - id: bflt_insptemp_u22
    type: b1le
  - id: bflt_lowpres_u21
    type: b1le
  - id: bflt_lowpres_u22
    type: b1le
  - id: bflt_highpres_u21
    type: b1le
  - id: bflt_highpres_u22
    type: b1le
  - id: bflt_diffpres_u2
    type: b1le
  - id: bflt_emergivt
    type: b1le
  - id: ig_rsv240
    type: b1le
  - id: ig_rsv241
    type: b1le
  - id: ig_rsv242
    type: b1le
  - id: bflt_vehtemp_u1
    type: b1le
  - id: ig_rsv251
    type: b1le
  - id: bflt_vehtemp_u2
    type: b1le
  - id: ig_rsv252
    type: b1le
  - id: bflt_airmon_u1
    type: b1le
  - id: bflt_airmon_u2
    type: b1le
  - id: bflt_currentmon
    type: b1le
  - id: bflt_tcms
    type: b1le
  - id: ig_rsv26
    type: u1
  - id: ig_rsv27
    type: u1
  - id: ig_rsv28
    type: u1
  - id: bflt_tempover
    type: b1le
  - id: bflt_powersupply_u1
    type: b1le
  - id: bflt_powersupply_u2
    type: b1le
  - id: bflt_exhaustfan
    type: b1le
  - id: bflt_exhaustval
    type: b1le
  - id: ig_rsv29
    type: b1le
  - id: ig_rsv30
    type: b1le
  - id: ig_rsv31
    type: b1le
  - id: ig_rsv32
    type: s2
  - id: ig_rsv33
    type: s2
  - id: fas_sys
    type: s2
  - id: ras_sys
    type: s2
  - id: tic
    type: s2
  - id: load
    type: s2
  - id: wrsv_42
    type: s2
  - id: tveh_1
    type: s2
  - id: humdity_1
    type: s2
  - id: tveh_2
    type: s2
  - id: humdity_2
    type: s2
  - id: aq_t_u1
    type: s2
  - id: aq_h_u1
    type: s2
  - id: aq_co2_u1
    type: s2
  - id: aq_tvoc_u1
    type: s2
  - id: aq_formald_u1
    type: s2
  - id: aq_pm2_5_u1
    type: s2
  - id: aq_pm10_u1
    type: s2
  - id: aq_rsv_u1
    type: s2
  - id: wmode_u1
    type: s2
  - id: presdiff_u1
    type: s2
  - id: fas_u1
    type: s2
  - id: ras_u1
    type: s2
  - id: fadpos_u1
    type: s2
  - id: radpos_u1
    type: s2
  - id: f_cp_u11
    type: s2
  - id: i_cp_u11
    type: s2
  - id: v_cp_u11
    type: s2
  - id: p_cp_u11
    type: s2
  - id: suckt_u11
    type: s2
  - id: suckp_u11
    type: s2
  - id: sp_u11
    type: s2
  - id: eevpos_u11
    type: s2
  - id: highpress_u11
    type: s2
  - id: sas_u11
    type: s2
  - id: ices_u11
    type: s2
  - id: f_cp_u12
    type: s2
  - id: i_cp_u12
    type: s2
  - id: v_cp_u12
    type: s2
  - id: p_cp_u12
    type: s2
  - id: suckt_u12
    type: s2
  - id: suckp_u12
    type: s2
  - id: sp_u12
    type: s2
  - id: eevpos_u12
    type: s2
  - id: highpress_u12
    type: s2
  - id: sas_u12
    type: s2
  - id: ices_u12
    type: s2
  - id: wrsv_124
    type: s2
  - id: aq_t_u2
    type: s2
  - id: aq_h_u2
    type: s2
  - id: aq_co2_u2
    type: s2
  - id: aq_tvoc_u2
    type: s2
  - id: aq_formald_u2
    type: s2
  - id: aq_pm2_5_u2
    type: s2
  - id: aq_pm10_u2
    type: s2
  - id: aq_rsv_u2
    type: s2
  - id: wmode_u2
    type: s2
  - id: presdiff_u2
    type: s2
  - id: fas_u2
    type: s2
  - id: ras_u2
    type: s2
  - id: fadpos_u2
    type: s2
  - id: radpos_u2
    type: s2
  - id: f_cp_u21
    type: s2
  - id: i_cp_u21
    type: s2
  - id: v_cp_u21
    type: s2
  - id: p_cp_u21
    type: s2
  - id: suckt_u21
    type: s2
  - id: suckp_u21
    type: s2
  - id: sp_u21
    type: s2
  - id: eevpos_u21
    type: s2
  - id: highpress_u21
    type: s2
  - id: sas_u21
    type: s2
  - id: ices_u21
    type: s2
  - id: f_cp_u22
    type: s2
  - id: i_cp_u22
    type: s2
  - id: v_cp_u22
    type: s2
  - id: p_cp_u22
    type: s2
  - id: suckt_u22
    type: s2
  - id: suckp_u22
    type: s2
  - id: sp_u22
    type: s2
  - id: eevpos_u22
    type: s2
  - id: highpress_u22
    type: s2
  - id: sas_u22
    type: s2
  - id: ices_u22
    type: s2
  - id: ig_rsv34
    type: s2
  - id: ig_rsv35
    type: s2
  - id: ig_rsv36
    type: s2
  - id: ig_rsv37
    type: s2
  - id: i_ef_u11
    type: s2
  - id: i_ef_u12
    type: s2
  - id: i_cf_u11
    type: s2
  - id: i_cf_u12
    type: s2
  - id: i_ef_u21
    type: s2
  - id: i_ef_u22
    type: s2
  - id: i_cf_u21
    type: s2
  - id: i_cf_u22
    type: s2
  - id: i_hvac_u1
    type: s2
  - id: i_hvac_u2
    type: s2
  - id: i_exufan
    type: s2
  - id: ig_rsv38
    type: s2
  - id: ig_rsv39
    type: s2
  - id: dwpower
    type: u4
  - id: dwemerg_op_tm
    type: u4
  - id: dwemerg_op_cnt
    type: u4
  - id: dwef_op_tm_u11
    type: u4
  - id: ig_rsv40
    type: u4
  - id: dwcf_op_tm_u11
    type: u4
  - id: ig_rsv41
    type: u4
  - id: dwcp_op_tm_u11
    type: u4
  - id: dwcp_op_tm_u12
    type: u4
  - id: ig_rsv42
    type: u4
  - id: ig_rsv43
    type: u4
  - id: dwfad_op_cnt_u1
    type: u4
  - id: dwrad_op_cnt_u1
    type: u4
  - id: dwef_op_cnt_u11
    type: u4
  - id: ig_rsv44
    type: u4
  - id: dwcf_op_cnt_u11
    type: u4
  - id: ig_rsv45
    type: u4
  - id: dwcp_op_cnt_u11
    type: u4
  - id: dwcp_op_cnt_u12
    type: u4
  - id: ig_rsv46
    type: u4
  - id: ig_rsv47
    type: u4
  - id: dwef_op_tm_u21
    type: u4
  - id: ig_rsv48
    type: u4
  - id: dwcf_op_tm_u21
    type: u4
  - id: ig_rsv49
    type: u4
  - id: dwcp_op_tm_u21
    type: u4
  - id: dwcp_op_tm_u22
    type: u4
  - id: ig_rsv50
    type: u4
  - id: ig_rsv51
    type: u4
  - id: dwfad_op_cnt_u2
    type: u4
  - id: dwrad_op_cnt_u2
    type: u4
  - id: dwef_op_cnt_u21
    type: u4
  - id: ig_rsv52
    type: u4
  - id: dwcf_op_cnt_u21
    type: u4
  - id: ig_rsv53
    type: u4
  - id: dwcp_op_cnt_u21
    type: u4
  - id: dwcp_op_cnt_u22
    type: u4
  - id: ig_rsv54
    type: u4
  - id: ig_rsv55
    type: u4
  - id: dwexufan_op_tm
    type: u4
  - id: dwexufan_op_cnt
    type: u4
  - id: dwdmpexu_op_cnt
    type: u4
  - id: ig_rsv56
    type: u4
  - id: ig_rsv57
    type: u4
  - id: ig_rsv58
    type: u4
  - id: ig_rsv59
    type: u4
  - id: ig_rsv60
    type: u4
  - id: ig_rsv61
    type: u4
  - id: ig_rsv62
    type: u4
  - id: ig_rsv63
    type: u4
  - id: ig_rsv64
    type: u4
  - id: ig_rsv65
    type: u4
  - id: ig_rsv66
    type: u4
  - id: ig_rsv67
    type: u4
  - id: ig_rsv68
    type: u4
  - id: dmp_exu_pos
    type: u2
  - id: start_station
    type: u2
  - id: terminal_station
    type: u2
  - id: cur_station
    type: u2
  - id: next_station
    type: u2