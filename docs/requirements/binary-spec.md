类型	序号	内容	偏移	字节数	说明
报文头	1   	报文头特征码	0	1	0x2C
            1	1	0x01
    2   	报文长度	2	2	整个报文长度，包括报文头、数据区及校验和
    3   	报文识别	4	1	源设备号,参见表1
            5	1	宿设备号,参见表1
            6	2	报文类型,参见具体的报文类型表
    4   	帧号	8	2	生命信号 (0~65535循环)
    5   	线路号	10	2	8
    6   	列车车型	12	2	10
    7   	列车号	14	4	列车编组编号
    8   	车厢号	18	1	1—6车
    9   	协议版本号	19	1	协议版本号  此版本号 0x01
    10      	预留	20	10	预留区域
数据区	11      	源设备时间	30	6	偏移由小到大依次是：年（实际年份减去2000以后的值），月（1-12），日（1-31），时（0-23），分（0-59），秒（0-59）
    12      	数据内容	36	N	N字节(协议内容)，如果源设备包含了各个子模块的数据，则协议内容中需要包含各个子模块接收数据的时间。
校验和	13      	校验和	36+N	2	不含自身的CRC校验和，从报文头开始至数据区结束，参见CRC校验节
                    
                    
备注：					
1.传输到地面的数据帧包含：报文头+空调数据+检验和					
2. 空调数据流首部有12个字节的文件头，包括车辆编码、车辆号、时间戳等信息；如果这些信息足够，协议里的报文头就可以不用解析了					
3. UDP协议实际上已经封装了校验功能；数据帧建议不用校验了，耗算力。					



    3.7 应用数据内容					
                        
序号	"字节
偏移"	位偏移	类型	数据名称(NB67）		值
1	0 		UNSIGNED8	Flag	dvc_flag	0
2	1 		UNSIGNED16	车辆编码	dvc_train_no	0
3	3 		UNSIGNED8	车厢号	dvc_carriage_no	3
4	4 		UNSIGNED8	年	dvc_year	24
5	5 		UNSIGNED8	月	dvc_month	7
6	6 		UNSIGNED8	日	dvc_day	19
7	7 		UNSIGNED8	时	dvc_hour	14
8	8 		UNSIGNED8	分	dvc_minute	36
9	9 		UNSIGNED8	秒	dvc_second	13
10	10 		UNSIGNED8	预留		
11	11 		UNSIGNED8	预留		
12	12	0	BOOL	通风机运行U1-1	cFBK_EF_U11	1
13		1	BOOL	预留		0
14		2	BOOL	冷凝风机运行U1-1	cFBK_CF_U11	1
15		3	BOOL	预留		0
16		4	BOOL	压缩机运行U11	cFBK_Comp_U11	1
17		5	BOOL	压缩机运行U12	cFBK_Comp_U12	1
18		6	BOOL	空气净化运行U1-1	cFBK_AP_U11	1
19		7	BOOL	预留		0
20	13	0	BOOL	通风机运行U2-1	cFBK_EF_U21	1
21		1	BOOL	预留		0
22		2	BOOL	冷凝风机运行U2-1	cFBK_CF_U21	1
23		3	BOOL	预留		0
24		4	BOOL	压缩机运行U21	cFBK_Comp_U21	1
25		5	BOOL	压缩机运行U22	cFBK_Comp_U22	1
26		6	BOOL	空气净化运行U2-1	cFBK_AP_U21	1
27		7	BOOL	预留		0
28	14	0	BOOL	机组1三相电源状态	cFBK_TPP_U1	1
29		1	BOOL	机组2三相电源状态	cFBK_TPP_U2	1
30		2	BOOL	紧急通风运行U1	cFBK_EV_U1	0
31		3	BOOL	紧急通风运行U2	cFBK_EV_U2	0
32		4	BOOL	废排紧急通风运行	cFBK_EWD	0
33		5	BOOL	预留		
34		6	BOOL	预留		
35		7	BOOL	预留		
36	15	0	BOOL	通风机过流故障U1-1	bOCFlt_EF_U11	0
37		1	BOOL	通风机过流故障U1-2	bOCFlt_EF_U12	0
38		2	BOOL	冷凝风机过流故障U1-1	bOCFlt_CF_U11	0
39		3	BOOL	冷凝风机过流故障U1-2	bOCFlt_CF_U12	0
40		4	BOOL	变频器保护故障U1-1	bFlt_VFD_U11	0
41		5	BOOL	变频器通讯故障U1-1	bFlt_VFD_COM_U11	0
42		6	BOOL	变频器保护故障U1-2	bFlt_VFD_U12	0
43		7	BOOL	变频器通讯故障U1-2	bFlt_VFD_COM_U12	0
44	16	0	BOOL	低压故障U1-1	bLPFlt_Comp_U11	0
45		1	BOOL	高压故障U1-1	bSCFlt_Comp_U11	0
46		2	BOOL	排气故障U1-1	bSCFlt_Vent_U11	0
47		3	BOOL	低压故障U1-2	bLPFlt_Comp_U12	0
48		4	BOOL	高压故障U1-2	bSCFlt_Comp_U12	0
49		5	BOOL	排气故障U1-2	bSCFlt_Vent_U11	0
50		6	BOOL	新风阀故障U1-1	bFlt_FAD_U11	1
51		7	BOOL	新风阀故障U1-2	bFlt_FAD_U12	1
52	17	0	BOOL	预留		0
53		1	BOOL	预留		0
54		2	BOOL	回风阀故障U1-1	bFlt_RAD_U11	1
55		3	BOOL	回风阀故障U1-2	bFlt_RAD_U12	1
56		4	BOOL	预留		0
57		5	BOOL	预留		0
58		6	BOOL	空气净化器故障U1-1	bFlt_AP_U11	0
59		7	BOOL	预留		0
60	18	0	BOOL	扩展模块通讯故障U1	bFlt_ExpBoard_U1	
61		1	BOOL	新风温度传感器故障U1	bFlt_FrsTemp_U1	
62		2	BOOL	回风温度传感器故障U1	bFlt_RntTemp_U1	
63		3	BOOL	送风温度传感器故障U1-1	bFlt_SplyTemp_U11	
64		4	BOOL	送风温度传感器故障U1-2	bFlt_SplyTemp_U12	
65		5	BOOL	盘管温度传感器故障U1-1	bFlt_CoilTemp_U11	
66		6	BOOL	盘管温度传感器故障U1-2	bFlt_CoilTemp_U12	
67		7	BOOL	吸气温度传感器故障U1-1	bFlt_InspTemp_U11	
68	19	0	BOOL	吸气温度传感器故障U1-2	bFlt_InspTemp_U12	
69		1	BOOL	低压压力传感器故障U1-1	bFlt_LowPres_U11	
70		2	BOOL	低压压力传感器故障U1-2	bFlt_LowPres_U12	
71		3	BOOL	高压压力传感器故障U1-1	bFlt_HighPres_U11	
72		4	BOOL	高压压力传感器故障U1-2	bFlt_HighPres_U12	
73		5	BOOL	压差传感器故障U1	bFlt_DiffPres_U1	1
74		6	BOOL	通风机过流故障U2-1	bOCFlt_EF_U21	
75		7	BOOL	通风机过流故障U2-2	bOCFlt_EF_U22	
76	20	0	BOOL	冷凝风机过流故障U2-1	bOCFlt_CF_U21	
77		1	BOOL	冷凝风机过流故障U2-2	bOCFlt_CF_U22	
78		2	BOOL	变频器保护故障U2-1	bFlt_VFD_U21	
79		3	BOOL	变频器通讯故障U2-1	bFlt_VFD_COM_U21	
80		4	BOOL	变频器保护故障U2-2	bFlt_VFD_U22	
81		5	BOOL	变频器通讯故障U2-2	bFlt_VFD_COM_U22	
82		6	BOOL	低压故障U2-1	bLPFlt_Comp_U21	
83		7	BOOL	高压故障U2-1	bSCFlt_Comp_U21	
84	21	0	BOOL	排气故障U2-1	bSCFlt_Vent_U21	
85		1	BOOL	低压故障U2-2	bLPFlt_Comp_U22	
86		2	BOOL	高压故障U2-2	bSCFlt_Comp_U22	
87		3	BOOL	排气故障U2-2	bSCFlt_Vent_U21	
88		4	BOOL	新风阀故障U2-1	bFlt_FAD_U21	1
89		5	BOOL	新风阀故障U2-2	bFlt_FAD_U22	1
90		6	BOOL	预留		
91		7	BOOL	预留		
92	22	0	BOOL	回风阀故障U2-1	bFlt_RAD_U21	1
93		1	BOOL	回风阀故障U2-2	bFlt_RAD_U22	1
94		2	BOOL	预留		
95		3	BOOL	预留		
96		4	BOOL	空气净化器故障U2-1	bFlt_AP_U21	
97		5	BOOL	预留		
98		6	BOOL	扩展模块通讯故障U2	bFlt_ExpBoard_U2	
99		7	BOOL	新风温度传感器故障U2	bFlt_FrsTemp_U2	
100	23	0	BOOL	回风温度传感器故障U2	bFlt_RntTemp_U2	
101		1	BOOL	送风温度传感器故障U2-1	bFlt_SplyTemp_U21	
102		2	BOOL	送风温度传感器故障U2-2	bFlt_SplyTemp_U22	
103		3	BOOL	盘管温度传感器故障U2-1	bFlt_CoilTemp_U21	
104		4	BOOL	盘管温度传感器故障U2-2	bFlt_CoilTemp_U22	
105		5	BOOL	吸气温度传感器故障U2-1	bFlt_InspTemp_U21	
106		6	BOOL	吸气温度传感器故障U2-2	bFlt_InspTemp_U22	
107		7	BOOL	低压压力传感器故障U2-1	bFlt_LowPres_U21	
108	24	0	BOOL	低压压力传感器故障U2-2	bFlt_LowPres_U22	
109		1	BOOL	高压压力传感器故障U2-1	bFlt_HighPres_U21	
110		2	BOOL	高压压力传感器故障U2-2	bFlt_HighPres_U22	
111		3	BOOL	压差传感器故障U2	bFlt_DiffPres_U2	1
112		4	BOOL	紧急逆变器故障	bFlt_EmergIVT	
113		5	BOOL	预留		
114		6	BOOL	预留		
115		7	BOOL	预留		
116	25	0	BOOL	车厢温度传感器1故障	bFlt_VehTemp_U1	
117		1	BOOL	预留		
118		2	BOOL	车厢温度传感器2故障	bFlt_VehTemp_U2	
119		3	BOOL	预留		
120		4	BOOL	机组1空气监测终端通讯故障	bFlt_AirMon_U1	1
121		5	BOOL	机组2空气监测终端通讯故障	bFlt_AirMon_U2	1
122		6	BOOL	电流监测单元通讯故障	bFlt_CurrentMon	1
123		7	BOOL	TCMS通讯故障	bFlt_TCMS	
124	26	0	BOOL	预留		
125		1	BOOL	预留		
126		2	BOOL	预留		
127		3	BOOL	预留		
128		4	BOOL	预留		
129		5	BOOL	预留		
130		6	BOOL	预留		
131		7	BOOL	预留		
132	27	0	BOOL	预留		
133		1	BOOL	预留		
134		2	BOOL	预留		
135		3	BOOL	预留		
136		4	BOOL	预留		
137		5	BOOL	预留		
138		6	BOOL	预留		
139		7	BOOL	预留		
140	28	0	BOOL	预留		
141		1	BOOL	预留		
142		2	BOOL	预留		
143		3	BOOL	预留		
144		4	BOOL	预留		
145		5	BOOL	预留		
146		6	BOOL	预留		
147		7	BOOL	预留		
148	29	0	BOOL	车厢温度超温预警	bFlt_TempOver	
149		1	BOOL	供电故障-U1	bFlt_PowerSupply_U1	
150		2	BOOL	供电故障-U2	bFlt_PowerSupply_U2	
151		3	BOOL	废排风机故障	bFlt_ExhaustFan	
152		4	BOOL	废排风阀故障	bFlt_ExhaustVal	1
153		5	BOOL	预留		
154		6	BOOL	预留		
155		7	BOOL	预留		
156	30		SIGNED16	预留		
157	32		SIGNED16	预留		
158	34		SIGNED16	新风温度-系统	fas_sys       	22
159	36		SIGNED16	回风温度-系统	ras_sys       	24.8
160	38		SIGNED16	目标温度	tic           	23
161	40		SIGNED16	载客量	load          	0
162	42		SIGNED16	预留	wRsv_42       	
163	44		SIGNED16	车厢温度-1	tVeh_1        	25.6
164	46		SIGNED16	预留	humdity_1     	
165	48		SIGNED16	车厢温度-2	tVeh_2        	25
166	50		SIGNED16	预留	humdity_2     	
167	52		SIGNED16	空气质量-温度-U1	AQ_t_u1       	0
168	54		SIGNED16	空气质量-湿度-U1	AQ_h_u1       	0
169	56		SIGNED16	空气质量-CO2-U1	AQ_CO2_u1     	0
170	58		SIGNED16	空气质量-TVOC-U1	AQ_TVOC_u1    	0
171	60		SIGNED16	空气质量-甲醛-U1	AQ_FORMALD_u1 	0
172	62		SIGNED16	空气质量-PM2.5-U1	AQ_PM2_5_u1   	0
173	64		SIGNED16	空气质量-PM10-U1	AQ_PM10_u1    	0
174	66		SIGNED16	预留	AQ_rsv_u1     	
175	68		SIGNED16	空调运行模式U1	wMode_u1      	2
176	70		SIGNED16	压差-U1	presDiff_u1   	3276.7
177	72		SIGNED16	新风温度-U1	fas_u1        	21.9
178	74		SIGNED16	回风温度-U1	ras_u1        	27.8
179	76		SIGNED16	新风阀开度-U1	fadpos_u1     	100
180	78		SIGNED16	回风阀开度-U1	radpos_u1     	48
181	80		SIGNED16	压缩机频率-U11	F_CP_u11      	0
182	82		SIGNED16	压缩机电流-U11	I_CP_u11      	0
183	84		SIGNED16	压缩机电压-U11	V_CP_u11      	0
184	86		SIGNED16	压缩机功率-U11	P_CP_u11      	0
185	88		SIGNED16	吸气温度-U11	suckT_u11     	21
186	90		SIGNED16	吸气压力-U11	suckP_u11     	2.5
187	92		SIGNED16	过热度-U11	sp_u11        	28.5
188	94		SIGNED16	电子膨胀阀开度-U11	eevpos_u11    	0
189	96		SIGNED16	高压压力-U11	highpress_u11 	1.2
190	98		SIGNED16	送风温度-U11	sas_u11       	-1.5
191	100		SIGNED16	盘管温度-U11	ices_u11      	-3.4
192	102		SIGNED16	压缩机频率-U12	F_CP_u12      	0
193	104		SIGNED16	压缩机电流-U12	I_CP_u12      	0
194	106		SIGNED16	压缩机电压-U12	V_CP_u12      	0
195	108		SIGNED16	压缩机功率-U12	P_CP_u12      	
196	110		SIGNED16	吸气温度-U12	suckT_u12     	
197	112		SIGNED16	吸气压力-U12	suckP_u12     	
198	114		SIGNED16	过热度-U12	sp_u12        	
199	116		SIGNED16	电子膨胀阀开度-U12	eevpos_u12    	
200	118		SIGNED16	高压压力-U12	highpress_u12 	
201	120		SIGNED16	送风温度-U12	sas_u12       	
202	122		SIGNED16	盘管温度-U11	ices_u12      	
203	124		SIGNED16	预留	wRsv_124	
204	126		SIGNED16	空气质量-温度-U2	AQ_t_u2      	
205	128		SIGNED16	空气质量-湿度-U2	AQ_h_u2      	
206	130		SIGNED16	空气质量-CO2-U2	AQ_CO2_u2    	
207	132		SIGNED16	空气质量-TVOC-U2	AQ_TVOC_u2   	
208	134		SIGNED16	空气质量-甲醛-U2	AQ_FORMALD_u2	
209	136		SIGNED16	空气质量-PM2.5-U2	AQ_PM2_5_u2  	
210	138		SIGNED16	空气质量-PM10-U2	AQ_PM10_u2   	
211	140		SIGNED16	预留	AQ_rsv_u2    	
212	142		SIGNED16	空调运行模式U2	wMode_u2     	
213	144		SIGNED16	压差-U2	presDiff_u2  	
214	146		SIGNED16	新风温度-U2	fas_u2       	
215	148		SIGNED16	回风温度-U2	ras_u2       	
216	150		SIGNED16	新风阀开度-U2	fadpos_u2    	
217	152		SIGNED16	回风阀开度-U2	radpos_u2    	
218	154		SIGNED16	压缩机频率-U21	F_CP_u21     	
219	156		SIGNED16	压缩机电流-U21	I_CP_u21     	
220	158		SIGNED16	压缩机电压-U21	V_CP_u21     	
221	160		SIGNED16	压缩机功率-U21	P_CP_u21     	
222	162		SIGNED16	吸气温度-U21	suckT_u21    	
223	164		SIGNED16	吸气压力-U21	suckP_u21    	
224	166		SIGNED16	过热度-U21	sp_u21       	
225	168		SIGNED16	电子膨胀阀开度-U21	eevpos_u21   	
226	170		SIGNED16	高压压力-U21	highpress_u21	
227	172		SIGNED16	送风温度-U21	sas_u21      	
228	174		SIGNED16	盘管温度-U21	ices_u21     	
229	176		SIGNED16	压缩机频率-U22	F_CP_u22     	
230	178		SIGNED16	压缩机电流-U22	I_CP_u22     	
231	180		SIGNED16	压缩机电压-U22	V_CP_u22     	
232	182		SIGNED16	压缩机功率-U22	P_CP_u22     	
233	184		SIGNED16	吸气温度-U22	suckT_u22    	
234	186		SIGNED16	吸气压力-U22	suckP_u22    	
235	188		SIGNED16	过热度-U22	sp_u22       	
236	190		SIGNED16	电子膨胀阀开度-U22	eevpos_u22   	
237	192		SIGNED16	高压压力-U22	highpress_u22	
238	194		SIGNED16	送风温度-U22	sas_u22      	
239	196		SIGNED16	盘管温度-U22	ices_u22     	
240	198		SIGNED16	预留		
241	200		SIGNED16	预留		
242	202		SIGNED16	预留		
243	204		SIGNED16	预留		
244	206		SIGNED16	通风机电流-U11	 I_EF_u11  	1.3
245	208		SIGNED16	通风机电流-U12	 I_EF_u12  	1.5
246	210		SIGNED16	冷凝风机电流-U11	 I_CF_u11  	2.3
247	212		SIGNED16	冷凝风机电流-U12	 I_CF_u12  	2.4
248	214		SIGNED16	通风机电流-U21	 I_EF_u21  	1.7
249	216		SIGNED16	通风机电流-U22	 I_EF_u22  	2.2
250	218		SIGNED16	冷凝风机电流-U21	 I_CF_u21  	2.5
251	220		SIGNED16	冷凝风机电流-U22	 I_CF_u22  	2.6
252	222		SIGNED16	预留	 I_HVAC_u1 	
253	224		SIGNED16	预留	 I_HVAC_u2 	
254	226		SIGNED16	废排风机电流	 I_EXUFan  	1.9
255	228		SIGNED16	预留		
256	230		SIGNED16	预留		
257	232		UNSIGNED32	空调机组能耗	dwPower	679889
258	236		UNSIGNED32	紧急逆变器累计运行时间	dwemerg_op_tm        	226
259	240		UNSIGNED32	紧急逆变器累计运行次数	dwemerg_op_cnt       	6
260	244		UNSIGNED32	通风机累计运行时间-U1-1	dwEF_op_tm_u11       	53334
261	248		UNSIGNED32	预留		
262	252		UNSIGNED32	冷凝风机累计运行时间-U1-1	dwCF_op_tm_u11       	
263	256		UNSIGNED32	预留		
264	260		UNSIGNED32	压缩机累计运行时间-U11	dwCP_op_tm_u11       	
265	264		UNSIGNED32	压缩机累计运行时间-U12	dwCP_op_tm_u12       	
266	268		UNSIGNED32	预留		
267	272		UNSIGNED32	预留		
268	276		UNSIGNED32	新风阀开关次数-U1	dwFAD_op_cnt_u1      	
269	280		UNSIGNED32	回风阀开关次数-U1	dwRAD_op_cnt_u1      	
270	284		UNSIGNED32	通风机累计开关次数-U1-1	dwEF_op_cnt_u11      	
271	288		UNSIGNED32	预留		
272	292		UNSIGNED32	冷凝风机累计开关次数-U1-1	dwCF_op_cnt_u11      	
273	296		UNSIGNED32	预留		
274	300		UNSIGNED32	压缩机累计开关次数-U11	dwCP_op_cnt_u11      	
275	304		UNSIGNED32	压缩机累计开关次数-U12	dwCP_op_cnt_u12      	
276	308		UNSIGNED32	预留		
277	312		UNSIGNED32	预留		
278	316		UNSIGNED32	通风机累计运行时间-U2-1	dwEF_op_tm_u21        	
279	320		UNSIGNED32	预留		
280	324		UNSIGNED32	冷凝风机累计运行时间-U2-1	dwCF_op_tm_u21        	
281	328		UNSIGNED32	预留		
282	332		UNSIGNED32	压缩机累计运行时间-U2-1	dwCP_op_tm_u21        	
283	336		UNSIGNED32	压缩机累计运行时间-U2-2	dwCP_op_tm_u22        	
284	340		UNSIGNED32	预留		
285	344		UNSIGNED32	预留		
286	348		UNSIGNED32	新风阀开关次数-U2	dwFAD_op_cnt_u2       	
287	352		UNSIGNED32	回风阀开关次数-U2	dwRAD_op_cnt_u2       	
288	356		UNSIGNED32	通风机累计开关次数-U2-1	dwEF_op_cnt_u21       	
289	360		UNSIGNED32	预留		
290	364		UNSIGNED32	冷凝风机累计开关次数-U2-1	dwCF_op_cnt_u21       	
291	368		UNSIGNED32	预留		
292	372		UNSIGNED32	压缩机累计开关次数-U2-1	dwCP_op_cnt_u21       	
293	376		UNSIGNED32	压缩机累计开关次数-U2-2	dwCP_op_cnt_u22       	
294	380		UNSIGNED32	预留		
295	384		UNSIGNED32	预留		
296	388		UNSIGNED32	废排风机累计运行时间	dwEXUFan_op_tm	
297	392		UNSIGNED32	废排风机累计开关次数	dwEXUFan_op_cnt	20
298	396		UNSIGNED32	废排风阀开关次数	dwDmpEXU_op_cnt	50
299	400		UNSIGNED32	预留		
300	404		UNSIGNED32	预留		
301	408		UNSIGNED32	预留		
302	412		UNSIGNED32	预留		
303	416		UNSIGNED32	预留		
304	420		UNSIGNED32	预留		
305	424		UNSIGNED32	预留		
306	428		UNSIGNED32	预留		
307	432		UNSIGNED32	预留		
308	436		UNSIGNED32	预留		
309	440		UNSIGNED32	预留		
310	444		UNSIGNED32	预留		
311	448		UNSIGNED32	预留		
312	452		UNSIGNED16	废排风阀开度	DmpEXU_pos	78
313	454		UNSIGNED16	起始站	Start_Station	291
314	456		UNSIGNED16	终点站	Terminal_Station	129
315	458		UNSIGNED16	当前站	Cur_Station	45
316	460		UNSIGNED16	下一站	Next_Station	66
317	462		UNSIGNED16	预留		
318	464		UNSIGNED16	预留		
319	466		UNSIGNED16	预留		
320	468		UNSIGNED16	预留		
321	470		UNSIGNED16	预留		
322	472		UNSIGNED16	预留		
323	474		UNSIGNED16	预留		
324	476		UNSIGNED16	预留		
325	478		UNSIGNED16	预留		
326	480		UNSIGNED16	预留		
327	482		UNSIGNED16	预留		
328	484		UNSIGNED16	预留		
329	486		UNSIGNED16	预留		
330	488		UNSIGNED16	预留		
331	490		UNSIGNED16	预留		
332	492		UNSIGNED16	预留		
333	494		UNSIGNED16	预留		
334	496		UNSIGNED16	预留		
335	498		UNSIGNED16	预留		