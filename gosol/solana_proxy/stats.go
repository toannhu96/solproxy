package solana_proxy

/*
import (
	"fmt"
	"strings"
	"time"

	"gosol/solana_proxy/throttle"

	"github.com/slawomir-pryczek/handler_socket2"
	"github.com/slawomir-pryczek/handler_socket2/hscommon"
)

func (this *SOLClient) GetThrottle() throttle.Throttle {
	_a, _b, _c := this._getThrottleStats()
	return throttle.Make(this.is_public_node, _a, _b, _c)
}

func (this *SOLClient) _getThrottleStats() (int, int, int) {

	stat_requests := 0
	stat_data_received := 0
	stat_requests_per_fn_max := 0

	// calculate requests
	_pos := this.stat_last_60_pos
	for i := 0; i < throttle.ThrottleConfig.Throttle_s_requests; i++ {
		stat_requests += this.stat_last_60[_pos].stat_done
		_pos-- // take current second into account
		if _pos < 0 {
			_pos = 59
		}
	}

	// calculate data received
	_pos = this.stat_last_60_pos
	for i := 0; i < throttle.ThrottleConfig.Throttle_s_data_received; i++ {
		stat_data_received += this.stat_last_60[_pos].stat_bytes_received
		_pos-- // take current second into account
		if _pos < 0 {
			_pos = 59
		}
	}

	// calculate top function number of calls
	requests_max_per_method := make(map[string]int)
	_pos = this.stat_last_60_pos
	for i := 0; i < throttle.ThrottleConfig.Throttle_s_requests_per_fn_max; i++ {
		for k, v := range this.stat_last_60[_pos].stat_request_by_fn {
			requests_max_per_method[k] += v
		}
		_pos-- // take current second into account
		if _pos < 0 {
			_pos = 59
		}
	}
	for _, v := range requests_max_per_method {
		if v > stat_requests_per_fn_max {
			stat_requests_per_fn_max = v
		}
	}

	return stat_requests, stat_requests_per_fn_max, stat_data_received
}

func (this *SOLClient) _statsAggr(seconds int) stat {

	s := stat{}
	_pos := this.stat_last_60_pos
	for i := 0; i < seconds; i++ {
		_pos--
		if _pos < 0 {
			_pos = 59
		}

		_tmp := this.stat_last_60[_pos%60]
		s.stat_done += _tmp.stat_done
		s.stat_error_json_decode += _tmp.stat_error_json_decode
		s.stat_error_json_marshal += _tmp.stat_error_json_marshal
		s.stat_error_req += _tmp.stat_error_req
		s.stat_error_resp += _tmp.stat_error_resp
		s.stat_error_resp_read += _tmp.stat_error_resp_read
		s.stat_ns_total += _tmp.stat_ns_total

		_tmp2 := make(map[string]int)
		for k, v := range _tmp.stat_request_by_fn {
			_tmp2[k] = _tmp2[k] + v
		}
		s.stat_request_by_fn = _tmp2
		s.stat_bytes_received += _tmp.stat_bytes_received
		s.stat_bytes_sent += _tmp.stat_bytes_sent
	}

	return s
}

func init() {

	start_time := time.Now().Unix()

	handler_socket2.StatusPluginRegister(func() (string, string) {

		_get_row := func(label string, s stat, time_running int, _addl ...string) []string {

			_req := fmt.Sprintf("%d", s.stat_done)
			_req_s := fmt.Sprintf("%.02f", float64(s.stat_done)/float64(time_running))
			_req_avg := fmt.Sprintf("%.02f ms", (float64(s.stat_ns_total)/float64(s.stat_done))/1000.0)

			_r := make([]string, 0, 10)
			_r = append(_r, label, _req, _req_s, _req_avg)
			_r = append(_r, _addl...)

			_r = append(_r, fmt.Sprintf("%d", s.stat_error_json_marshal))
			_r = append(_r, fmt.Sprintf("%d", s.stat_error_req))
			_r = append(_r, fmt.Sprintf("%d", s.stat_error_resp))
			_r = append(_r, fmt.Sprintf("%d", s.stat_error_resp_read))
			_r = append(_r, fmt.Sprintf("%d", s.stat_error_json_decode))

			_r = append(_r, fmt.Sprintf("%.02fMB", float64(s.stat_bytes_sent)/1000/1000))
			_r = append(_r, fmt.Sprintf("%.02fMB", float64(s.stat_bytes_received)/1000/1000))
			return _r
		}

		time_running := time.Now().Unix() - start_time
		mu.Lock()

		status := ""
		for _, v := range clients {

			_a, _b, _c := v._getThrottleStats()
			thr := throttle.Make(v.is_public_node, _a, _b, _c)

			_t := "Private"
			if v.is_public_node {
				_t = "Public"
			}

			_tmp := thr.GetThrottledStatus()
			color := "#44aa44"
			if _tmp["is_throttled"].(bool) {
				color = "#aa4444"
			}

			node_stats := "<b style='background: #FF7777!important'>Broken / Disabled!</b>"
			if v.is_disabled {
				node_stats = "<b style='background: #77FF77!important'>Running</b>"
			}

			__t, __t2 := v.IsAlive()
			node_stats = fmt.Sprintf("Node status: %s, Based on current stats (%d seconds) next alive status is: <b>%v</b> (using %d requests)<br>",
				node_stats, probe_isalive_seconds, __t, __t2)

			throttle_stats := fmt.Sprintf("Throttle settings: <b style='color:%s'>%s</b>\n", color, _tmp["throttled_comment"])
			for k, v := range _tmp {
				if strings.Index(k, "throttle_") == -1 {
					continue
				}
				throttle_stats += fmt.Sprintf("<b>%s</b>: %s\n", k, v)
			}

			table := hscommon.NewTableGen("Time", "Requests", "Req/s", "Avg Time", "First Block",
				"Err JM", "Err Req", "Err Resp", "Err RResp", "Err Decode", "Sent", "Received")
			table.SetClass("tab sol")

			status += "\n"
			status += _t + " Node Endpoint: " + v.endpoint + " <i>v" + v.version + "</i> ... Requests running now: " + fmt.Sprintf("%d", v.stat_running) + "\n"
			status += node_stats
			status += throttle_stats

			table.AddRow(_get_row("Last 10s", v._statsAggr(10), 10, "-")...)
			table.AddRow(_get_row("Last 60s", v._statsAggr(59), 59, "-")...)

			_fb := fmt.Sprintf("%d", v.first_available_block)
			table.AddRow(_get_row("Total", v.stat_total, int(time_running), _fb)...)
			status += table.Render()
		}
		mu.Unlock()

		info := "This section represents individual SOLANA nodes, with number of requests and errors\n"
		info += "<b>Err JM</b> - Json Marshall error. We were unable to build JSON payload required for your request\n"
		info += "<b>Err Req</b> - Request Error. We were unable to send request to host\n"
		info += "<b>Err Resp</b> - Response Error. We were unable to get server response\n"
		info += "<b>Err RResp</b> - Response Reading Error. We were unable to read server response\n"
		info += "<b>Err Decode</b> - Json Decode Error. We were unable read received JSON\n"
		return "Solana Proxy", "<pre>" + info + status + "</pre>"
	})
}
*/
