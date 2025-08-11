package web

import (
	"ConfigProbe/pkg/v2rayprobe/litespeedtest/config"
	"ConfigProbe/pkg/v2rayprobe/litespeedtest/download"
	"ConfigProbe/pkg/v2rayprobe/litespeedtest/request"
	"ConfigProbe/pkg/v2rayprobe/litespeedtest/utils"
	"ConfigProbe/pkg/v2rayprobe/litespeedtest/web/render"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidData = errors.New("invalid data")
	regProfile     = regexp.MustCompile(`((?i)vmess://(\S+?)@(\S+?):([0-9]{2,5})/([?#][^\s]+))|((?i)vmess://[a-zA-Z0-9+_/=-]+([?#][^\s]+)?)|((?i)ssr://[a-zA-Z0-9+_/=-]+)|((?i)(vless|ss|trojan)://(\S+?)@(\S+?):([0-9]{2,5})/?([?#][^\s]+))|((?i)(ss)://[a-zA-Z0-9+_/=-]+([?#][^\s]+))`)
)

const (
	PIC_BASE64 = iota
	PIC_PATH
	PIC_NONE
	JSON_OUTPUT
	TEXT_OUTPUT
)

type PAESE_TYPE int

const (
	PARSE_ANY PAESE_TYPE = iota
	PARSE_URL
	PARSE_FILE
	PARSE_BASE64
	PARSE_CLASH
	PARSE_PROFILE
)

const (
	remarkPrefix    = "SubFinder"
	MinCountToWrite = 10
)

// support proxy
// concurrency setting
// as subscription server
// profiles filter
// clash to vmess local subscription
func getSubscriptionLinks(link string) ([]string, error) {
	c := http.Client{
		Timeout: 20 * time.Second,
	}
	resp, err := c.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if isYamlFile(link) {
		return scanClashProxies(resp.Body, true)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	dataStr := string(data)
	msg, err := utils.DecodeB64(dataStr)
	if err != nil {
		if strings.Contains(dataStr, "proxies:") {
			return parseClash(dataStr)
		} else if strings.Contains(dataStr, "vmess://") ||
			strings.Contains(dataStr, "trojan://") ||
			strings.Contains(dataStr, "ssr://") ||
			strings.Contains(dataStr, "ss://") {
			return parseProfiles(dataStr)
		} else {
			return []string{}, err
		}
	}
	return ParseLinks(msg)
}

type parseFunc func(string) ([]string, error)

type ParseOption struct {
	Type PAESE_TYPE
}

// api
func ParseLinks(message string) ([]string, error) {
	opt := ParseOption{Type: PARSE_ANY}
	return ParseLinksWithOption(message, opt)
}

// api
func ParseLinksWithOption(message string, opt ParseOption) ([]string, error) {
	// matched, err := regexp.MatchString(`^(?:https?:\/\/)(?:[^@\/\n]+@)?(?:www\.)?([^:\/\n]+)`, message)
	if opt.Type == PARSE_URL || utils.IsUrl(message) {
		log.Println(message)
		return getSubscriptionLinks(message)
	}
	// check is file path
	if opt.Type == PARSE_FILE || utils.IsFilePath(message) {
		return parseFile(message)
	}
	if opt.Type == PARSE_BASE64 {
		return parseBase64(message)
	}
	if opt.Type == PARSE_CLASH {
		return parseClash(message)
	}
	if opt.Type == PARSE_PROFILE {
		return parseProfiles(message)
	}
	var links []string
	var err error
	for _, fn := range []parseFunc{parseProfiles, parseBase64, parseClash, parseFile} {
		links, err = fn(message)
		if err == nil && len(links) > 0 {
			break
		}
	}
	return links, err
}

func parseProfiles(data string) ([]string, error) {
	// encodeed url
	links := strings.Split(data, "\n")
	if len(links) > 1 {
		for i, link := range links {
			if l, err := url.Parse(link); err == nil {
				if query, err := url.QueryUnescape(l.RawQuery); err == nil && query == l.RawQuery {
					links[i] = l.String()
				}
			}
		}
		data = strings.Join(links, "\n")
	}
	// reg := regexp.MustCompile(`((?i)vmess://(\S+?)@(\S+?):([0-9]{2,5})/([?#][^\s]+))|((?i)vmess://[a-zA-Z0-9+_/=-]+([?#][^\s]+)?)|((?i)ssr://[a-zA-Z0-9+_/=-]+)|((?i)(vless|ss|trojan)://(\S+?)@(\S+?):([0-9]{2,5})([?#][^\s]+))|((?i)(ss)://[a-zA-Z0-9+_/=-]+([?#][^\s]+))`)
	matches := regProfile.FindAllStringSubmatch(data, -1)
	linksLen, matchesLen := len(links), len(matches)
	if linksLen < matchesLen {
		links = make([]string, matchesLen)
	} else if linksLen > matchesLen {
		links = links[:len(matches)]
	}
	for index, match := range matches {
		link := match[0]
		if config.RegShadowrocketVmess.MatchString(link) {
			if l, err := config.ShadowrocketLinkToVmessLink(link); err == nil {
				link = l
			}
		}
		links[index] = link
	}
	return links, nil
}

func parseBase64(data string) ([]string, error) {
	msg, err := utils.DecodeB64(data)
	if err != nil {
		return nil, err
	}
	return parseProfiles(msg)
}

func parseClash(data string) ([]string, error) {
	cc, err := config.ParseClash(utils.UnsafeGetBytes(data))
	if err != nil {
		return parseClashProxies(data)
	}
	return cc.Proxies, nil
}

// split to new line
func parseClashProxies(input string) ([]string, error) {

	if !strings.Contains(input, "{") {
		return []string{}, nil
	}
	return scanClashProxies(strings.NewReader(input), true)
}

func scanClashProxies(r io.Reader, greedy bool) ([]string, error) {
	proxiesStart := false
	var data []byte
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		b := scanner.Bytes()
		trimLine := strings.TrimSpace(string(b))
		if trimLine == "proxy-groups:" || trimLine == "rules:" || trimLine == "Proxy Group:" {
			break
		}
		if !proxiesStart && (trimLine == "proxies:" || trimLine == "Proxy:") {
			proxiesStart = true
			b = []byte("proxies:")
		}
		if proxiesStart {
			if _, err := config.ParseBaseProxy(trimLine); err != nil {
				continue
			}
			data = append(data, b...)
			data = append(data, byte('\n'))
		}
	}
	// fmt.Println(string(data))
	return parseClashByte(data)
}

func parseClashFileByLine(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return scanClashProxies(file, false)
}

func parseClashByte(data []byte) ([]string, error) {
	cc, err := config.ParseClash(data)
	if err != nil {
		return nil, err
	}
	return cc.Proxies, nil
}

func parseFile(filepath string) ([]string, error) {
	filepath = strings.TrimSpace(filepath)
	if _, err := os.Stat(filepath); err != nil {
		return nil, err
	}
	// clash
	if isYamlFile(filepath) {
		return parseClashFileByLine(filepath)
	}
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	links, err := parseBase64(string(data))
	if err != nil && len(data) > 2048 {
		preview := string(data[:2048])
		if strings.Contains(preview, "proxies:") {
			return scanClashProxies(bytes.NewReader(data), true)
		}
		if strings.Contains(preview, "vmess://") ||
			strings.Contains(preview, "trojan://") ||
			strings.Contains(preview, "ssr://") ||
			strings.Contains(preview, "ss://") {
			return parseProfiles(string(data))
		}
	}
	return links, err
}

func parseOptions(message string) (*ProfileTestOptions, error) {
	opts := strings.Split(message, "^")
	if len(opts) < 7 {
		return nil, ErrInvalidData
	}
	groupName := opts[0]
	if groupName == "?empty?" || groupName == "" {
		groupName = "Default"
	}
	concurrency, err := strconv.Atoi(opts[5])
	if err != nil {
		return nil, err
	}
	if concurrency < 1 {
		concurrency = 1
	}
	timeout, err := strconv.Atoi(opts[6])
	if err != nil {
		return nil, err
	}
	if timeout < 20 {
		timeout = 20
	}
	testOpt := &ProfileTestOptions{
		GroupName:     groupName,
		SpeedTestMode: opts[1],
		PingMethod:    opts[2],
		SortMethod:    opts[3],
		Concurrency:   concurrency,
		TestMode:      ALLTEST,
		Timeout:       time.Duration(timeout) * time.Second,
	}
	return testOpt, nil
}

const (
	SpeedOnly = "speedonly"
	PingOnly  = "pingonly"
	ALLTEST   = iota
	RETEST
)

type ProfileTestOptions struct {
	GroupName       string        `json:"group"`
	SpeedTestMode   string        `json:"speedtestMode"` // speedonly pingonly all
	PingMethod      string        `json:"pingMethod"`    // googleping
	SortMethod      string        `json:"sortMethod"`    // speed rspeed ping rping
	Concurrency     int           `json:"concurrency"`
	TestMode        int           `json:"testMode"` // 2: ALLTEST 3: RETEST
	TestIDs         []int         `json:"testids"`
	Timeout         time.Duration `json:"timeout"`
	Links           []string      `json:"links"`
	Subscription    string        `json:"subscription"`
	Language        string        `json:"language"`
	FontSize        int           `json:"fontSize"`
	Theme           string        `json:"theme"`
	Unique          bool          `json:"unique"`
	GeneratePicMode int           `json:"generatePicMode"` // 0: base64 1:pic path 2: no pic 3: json @deprecated use outputMode
	OutputMode      int           `json:"outputMode"`
}

type JSONOutput struct {
	Nodes        []render.Node      `json:"nodes"`
	Options      ProfileTestOptions `json:"options"`
	Traffic      int64              `json:"traffic"`
	Duration     string             `json:"duration"`
	SuccessCount int                `json:"successCount"`
	LinksCount   int                `json:"linksCount"`
}

func parseMessage(message []byte) ([]string, *ProfileTestOptions, error) {
	options := &ProfileTestOptions{}
	err := json.Unmarshal(message, options)
	if err != nil {
		return nil, nil, err
	}
	options.Timeout = time.Duration(int(options.Timeout)) * time.Second
	if options.GroupName == "?empty?" || options.GroupName == "" {
		options.GroupName = "Default"
	}
	if options.Timeout < 8 {
		options.Timeout = 8
	}
	if options.Concurrency < 1 {
		options.Concurrency = 1
	}
	if options.TestMode == RETEST {
		return options.Links, options, nil
	}
	options.TestMode = ALLTEST
	links, err := ParseLinks(options.Subscription)
	if err != nil {
		return nil, nil, err
	}
	return links, options, nil
}

func parseRetestMessage(message []byte) ([]string, *ProfileTestOptions, error) {
	options := &ProfileTestOptions{}
	err := json.Unmarshal(message, options)
	if err != nil {
		return nil, nil, err
	}
	if options.TestMode != RETEST {
		return nil, nil, errors.New("not retest mode")
	}
	options.TestMode = RETEST
	options.Timeout = time.Duration(int(options.Timeout)) * time.Second
	if options.GroupName == "?empty?" || options.GroupName == "" {
		options.GroupName = "Default"
	}
	if options.Timeout < 20 {
		options.Timeout = 20
	}
	if options.Concurrency < 1 {
		options.Concurrency = 1
	}
	return options.Links, options, nil
}

type MessageWriter interface {
	WriteMessage(messageType int, data []byte) error
}

type OutputMessageWriter struct {
}

func (p *OutputMessageWriter) WriteMessage(messageType int, data []byte) error {
	log.Println(string(data))
	return nil
}

type EmptyMessageWriter struct {
}

func (w *EmptyMessageWriter) WriteMessage(messageType int, data []byte) error {
	return nil
}

type ProfileTest struct {
	Writer      MessageWriter
	Options     *ProfileTestOptions
	MessageType int
	Links       []string
	mu          sync.Mutex
	wg          sync.WaitGroup // wait for all to finish
}

func (p *ProfileTest) WriteMessage(data []byte) error {
	var err error
	if p.Writer != nil {
		p.mu.Lock()
		err = p.Writer.WriteMessage(p.MessageType, data)
		p.mu.Unlock()
	}
	return err
}

func (p *ProfileTest) WriteString(data string) error {
	b := []byte(data)
	return p.WriteMessage(b)
}

// api
// render.Node contain the final test result
func (p *ProfileTest) TestAll(ctx context.Context, trafficChan chan<- int64) (chan render.Node, error) {
	links := p.Links
	linksCount := len(links)
	if linksCount < 1 {
		return nil, fmt.Errorf("profile not found")
	}
	nodeChan := make(chan render.Node, linksCount)
	go func(context.Context) {
		guard := make(chan int, p.Options.Concurrency)
		for i := range links {
			p.wg.Add(1)
			id := i
			link := links[i]
			select {
			case guard <- i:
				go func(id int, link string, c <-chan int, nodeChan chan<- render.Node) {
					p.testOne(ctx, id, link, nodeChan, trafficChan)
					<-c
				}(id, link, guard, nodeChan)
			case <-ctx.Done():
				return
			}
		}
		// p.wg.Wait()
		// if trafficChan != nil {
		// 	close(trafficChan)
		// }
	}(ctx)
	return nodeChan, nil
}

// Deprecated: use testAllBatch instead
func (p *ProfileTest) testAll(ctx context.Context) (render.Nodes, error) {
	linksCount := len(p.Links)
	if linksCount < 1 {
		p.WriteString(SPEEDTEST_ERROR_NONODES)
		return nil, fmt.Errorf("no profile found")
	}
	start := time.Now()
	p.WriteMessage(getMsgByte(-1, "started"))
	// for i := range p.Links {
	// 	p.WriteMessage(gotserverMsg(i, p.Links[i], p.Options.GroupName))
	// }
	step := 9
	if linksCount > 200 {
		step = linksCount / 20
		if step > 50 {
			step = 50
		}
	}
	for i := 0; i < linksCount; {
		end := i + step
		if end > linksCount {
			end = linksCount
		}
		links := p.Links[i:end]
		msg := gotserversMsg(i, links, p.Options.GroupName)
		p.WriteMessage(msg)
		i += step
	}
	guard := make(chan int, p.Options.Concurrency)
	nodeChan := make(chan render.Node, linksCount)

	nodes := make(render.Nodes, linksCount)
	for i := range p.Links {
		p.wg.Add(1)
		id := i
		link := ""
		if len(p.Options.TestIDs) > 0 && len(p.Options.Links) > 0 {
			id = p.Options.TestIDs[i]
			link = p.Options.Links[i]
		}
		select {
		case guard <- i:
			go func(id int, link string, c <-chan int, nodeChan chan<- render.Node) {
				seconds := time.Duration(2) * p.Options.Timeout
				timeoutCtx, cacnel := context.WithTimeout(context.Background(), seconds)
				defer cacnel()

				p.testOne(timeoutCtx, id, link, nodeChan, nil)
				_ = p.WriteMessage(getMsgByte(id, "endone"))
				<-c
			}(id, link, guard, nodeChan)
		case <-ctx.Done():
			return nil, nil
		}
	}
	p.wg.Wait()
	p.WriteMessage(getMsgByte(-1, "eof"))
	duration := FormatDuration(time.Since(start))
	// draw png
	successCount := 0
	var traffic int64 = 0
	for i := 0; i < linksCount; i++ {
		node := <-nodeChan
		node.Link = p.Links[node.Id]
		nodes[node.Id] = node
		traffic += node.Traffic
		if node.IsOk {
			successCount += 1
		}
	}
	close(nodeChan)

	if p.Options.OutputMode == PIC_NONE {
		return nodes, nil
	}

	// sort nodes
	nodes.Sort(p.Options.SortMethod)
	nodes.ChangeRemarks(remarkPrefix)
	// save json
	if len(nodes) >= MinCountToWrite {
		if p.Options.OutputMode == JSON_OUTPUT {
			p.saveJSON(nodes, traffic, duration, successCount, linksCount)
		} else if p.Options.OutputMode == TEXT_OUTPUT {
			p.saveText(nodes)
		}
	}
	return nodes, nil
}

func (p *ProfileTest) testAllBatch(ctx context.Context, workerPoolSize int) (render.Nodes, error) {
	linksCount := len(p.Links)
	log.Printf("Count: %d", linksCount)
	if linksCount < 1 {
		p.WriteString(SPEEDTEST_ERROR_NONODES)
		return nil, fmt.Errorf("no profile found")
	}
	start := time.Now()
	p.WriteMessage(getMsgByte(-1, "started"))

	nodes := make(render.Nodes, linksCount)
	var traffic int64 = 0
	successCount := 0

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	// Worker pool configuration
	maxWorkers := workerPoolSize // Adjust this number based on your needs
	sem := make(chan struct{}, maxWorkers)

	for i := 0; i < linksCount; i++ {
		id := i
		link := ""
		if len(p.Options.TestIDs) > 0 && len(p.Options.Links) > 0 {
			id = p.Options.TestIDs[i]
			link = p.Options.Links[i]
		}

		wg.Add(1)
		go func(id int, link string) {
			defer wg.Done()
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Create timeout context for each test
			seconds := time.Duration(2) * p.Options.Timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, seconds)
			defer cancel()

			node, err := p.testOneSync(timeoutCtx, id, link)
			if err != nil {
				p.WriteMessage(getMsgByte(id, "endone"))
			}

			// Update node properties
			node.Link = p.Links[node.Id]

			// Synchronized write to shared data
			mu.Lock()
			defer mu.Unlock()

			nodes[id] = node
			traffic += node.Traffic
			if node.IsOk {
				successCount++
			}
		}(id, link)
	}

	wg.Wait()

	duration := FormatDuration(time.Since(start))
	// Rest of the output handling remains the same
	if p.Options.OutputMode == PIC_NONE {
		return nodes, nil
	}
	nodes.Sort(p.Options.SortMethod)
	nodes.ChangeRemarks(remarkPrefix)
	if len(nodes) >= MinCountToWrite {
		if p.Options.OutputMode == JSON_OUTPUT {
			p.saveJSON(nodes, traffic, duration, successCount, linksCount)
		} else if p.Options.OutputMode == TEXT_OUTPUT {
			p.saveText(nodes)
		}
	}
	return nodes, nil
}

func (p *ProfileTest) saveJSON(nodes render.Nodes, traffic int64, duration string, successCount int, linksCount int) error {
	jsonOutput := JSONOutput{
		Nodes:        nodes,
		Options:      *p.Options,
		Traffic:      traffic,
		Duration:     duration,
		SuccessCount: successCount,
		LinksCount:   linksCount,
	}
	data, err := json.MarshalIndent(&jsonOutput, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile("output.json", data, 0644)
}

func (p *ProfileTest) saveText(nodes render.Nodes) error {
	var links []string
	for _, node := range nodes {
		if node.Ping != "0" || node.AvgSpeed > 0 || node.MaxSpeed > 0 {
			links = append(links, node.Link)
		}
	}
	data := []byte(strings.Join(links, "\n"))
	return ioutil.WriteFile("data/output.txt", data, 0644)
}

func (p *ProfileTest) testOne(ctx context.Context, index int, link string, nodeChan chan<- render.Node, trafficChan chan<- int64) error {
	// panic
	defer p.wg.Done()
	if link == "" {
		link = p.Links[index]
		link = strings.SplitN(link, "^", 2)[0]
	}
	cfg, err := config.Link2Config(link)
	if err != nil {
		return err
	}
	remarks := cfg.Remarks
	if err != nil || remarks == "" {
		remarks = fmt.Sprintf("Profile %d", index)
	}
	protocol := cfg.Protocol
	if (cfg.Protocol == "vmess" || cfg.Protocol == "trojan") && cfg.Net != "" {
		protocol = fmt.Sprintf("%s/%s", cfg.Protocol, cfg.Net)
	}

	var elapse int64 = 0
	if p.Options.SpeedTestMode != SpeedOnly {
		elapse, err = p.pingLink(index, link)
		log.Printf("%d %s elapse: %dms", index, remarks, elapse)

		if err != nil {
			node := render.Node{
				Id:       index,
				Group:    p.Options.GroupName,
				Remarks:  remarks,
				Protocol: protocol,
				Ping:     fmt.Sprintf("%d", elapse),
				AvgSpeed: 0,
				MaxSpeed: 0,
				IsOk:     elapse > 0,
			}
			nodeChan <- node
			return err
		}
	}
	err = p.WriteMessage(getMsgByte(index, "startspeed"))
	ch := make(chan int64, 1)
	startCh := make(chan time.Time, 1)
	defer close(ch)
	go func(ch <-chan int64, startChan <-chan time.Time) {
		var max int64
		var sum int64
		var avg int64
		start := time.Now()
	Loop:
		for {
			select {
			case speed, ok := <-ch:
				if !ok || speed < 0 {
					break Loop
				}
				sum += speed
				duration := float64(time.Since(start)/time.Millisecond) / float64(1000)
				avg = int64(float64(sum) / duration)
				if max < speed {
					max = speed
				}
				log.Printf("%d %s recv: %s", index, remarks, download.ByteCountIEC(speed))
				err = p.WriteMessage(getMsgByte(index, "gotspeed", avg, max, speed))
				if trafficChan != nil {
					trafficChan <- speed
				}
			case s := <-startChan:
				start = s
			case <-ctx.Done():
				log.Printf("index %d done!", index)
				break Loop
			}
		}
		node := render.Node{
			Id:       index,
			Group:    p.Options.GroupName,
			Remarks:  remarks,
			Protocol: protocol,
			Ping:     fmt.Sprintf("%d", elapse),
			AvgSpeed: avg,
			MaxSpeed: max,
			IsOk:     true,
			Traffic:  sum,
		}
		nodeChan <- node
	}(ch, startCh)
	speed, err := download.Download(link, p.Options.Timeout, p.Options.Timeout, ch, startCh)
	// speed, err := download.DownloadRange(link, 2, p.Options.Timeout, p.Options.Timeout, ch, startCh)
	if speed < 1 {
		p.WriteMessage(getMsgByte(index, "gotspeed", -1, -1, 0))
	}
	return err
}

func (p *ProfileTest) testOneSync(ctx context.Context, index int, link string) (render.Node, error) {
	var node render.Node

	// Determine the link to test
	if link == "" {
		link = p.Links[index]
		link = strings.SplitN(link, "^", 2)[0]
	}
	cfg, err := config.Link2Config(link)
	if err != nil {
		log.Printf("Link %d: not working (configuration error: %v)", index, err)
		return node, err
	}
	remarks := cfg.Remarks
	if remarks == "" {
		remarks = fmt.Sprintf("Profile %d", index)
	}
	protocol := cfg.Protocol
	if (cfg.Protocol == "vmess" || cfg.Protocol == "trojan") && cfg.Net != "" {
		protocol = fmt.Sprintf("%s/%s", cfg.Protocol, cfg.Net)
	}
	log.Printf("Testing link %d (%s)", index, remarks)

	var elapse int64 = 0
	if p.Options.SpeedTestMode != SpeedOnly {
		elapse, err = p.pingLink(index, link)
		log.Printf("Ping for link %d (%s): %dms", index, remarks, elapse)
		if err != nil {
			node = render.Node{
				Id:       index,
				Group:    p.Options.GroupName,
				Remarks:  remarks,
				Protocol: protocol,
				Ping:     fmt.Sprintf("%d", elapse),
				AvgSpeed: 0,
				MaxSpeed: 0,
				IsOk:     elapse > 0,
			}
			log.Printf("Link %d (%s): not working (ping error: %v)", index, remarks, err)
			return node, err
		}
	}

	_ = p.WriteMessage(getMsgByte(index, "startspeed"))
	// Call download synchronously
	speed, err := download.Download(link, p.Options.Timeout, p.Options.Timeout, nil, nil)
	if speed < 1 {
		p.WriteMessage(getMsgByte(index, "gotspeed", -1, -1, 0))
	}

	// Create the node based on the download result.
	node = render.Node{
		Id:       index,
		Group:    p.Options.GroupName,
		Remarks:  remarks,
		Protocol: protocol,
		Ping:     fmt.Sprintf("%d", elapse),
		AvgSpeed: speed, // Using the final speed as both average and max.
		MaxSpeed: speed,
		IsOk:     err == nil && speed > 0,
		Traffic:  speed,
	}

	// Final log: one line per link indicating working or not working.
	if err != nil || speed < 1 {
		log.Printf("Link %d (%s): not working (speed: %d, error: %v)", index, remarks, speed, err)
	} else {
		log.Printf("Link %d (%s): working (speed: %d)", index, remarks, speed)
	}

	return node, err
}

func (p *ProfileTest) pingLink(index int, link string) (int64, error) {
	if link == "" {
		link = p.Links[index]
	}
	p.WriteMessage(getMsgByte(index, "startping"))
	elapse, err := request.PingLink(link, 2)
	p.WriteMessage(getMsgByte(index, "gotping", elapse))
	if elapse < 1 {
		p.WriteMessage(getMsgByte(index, "gotspeed", -1, -1, 0))
		return 0, err
	}
	if p.Options.SpeedTestMode == PingOnly {
		p.WriteMessage(getMsgByte(index, "gotspeed", -1, -1, 0))
		return elapse, errors.New(PingOnly)
	}
	return elapse, err
}

func FormatDuration(duration time.Duration) string {
	h := duration / time.Hour
	duration -= h * time.Hour
	m := duration / time.Minute
	duration -= m * time.Minute
	s := duration / time.Second
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

func isYamlFile(filePath string) bool {
	return strings.HasSuffix(filePath, ".yaml") || strings.HasSuffix(filePath, ".yml")
}
