package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
)

var modelMapping = map[string]string{
    "deepseek-r1": "ep-20250218144306-sshj5",
    "deepseek-v3": "ep-20250218170310-pvmdb",
    "deepseek-r1-distill-qwen-7b": "ep-20250218182157-7rdv4",
    "deepseek-r1-distill-qwen-32b": "ep-20250218182233-vxkqv"
}

type ChatRequest struct {
    Model    string                   `json:"model" binding:"required"`
    Messages []map[string]interface{} `json:"messages" binding:"required"`
    Stream   bool                     `json:"stream"`
}

func fangzou(apiKey, model string, messages []map[string]interface{}, stream bool, extraParams map[string]interface{}) (interface{}, error) {
    targetModel := modelMapping[model]
    if targetModel == "" {
        targetModel = modelMapping["deepseek-r1"]
    }

    baseURL := "https://ark.cn-beijing.volces.com/api/v3/chat/completions"

    payload := map[string]interface{}{
        "model":    targetModel,
        "messages": messages,
        "stream":   stream,
    }
    for k, v := range extraParams {
        payload[k] = v
    }

    payloadBytes, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("marshal payload failed: %w", err)
    }

    req, err := http.NewRequest("POST", baseURL, bytes.NewReader(payloadBytes))
    if err != nil {
        return nil, fmt.Errorf("create request failed: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)

    client := &http.Client{
        Timeout: 1800 * time.Second,
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("backend request failed: %w", err)
    }

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        resp.Body.Close()
        return nil, fmt.Errorf("backend error (%d): %s", resp.StatusCode, string(body))
    }

    if stream {
        return resp.Body, nil // 返回未关闭的响应体
    }

    defer resp.Body.Close()
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response failed: %w", err)
    }
    return result, nil
}

func handleRequest(c *gin.Context) {
    // 捕获并记录原始请求体
    bodyBytes, _ := io.ReadAll(c.Request.Body)
    c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

    fmt.Printf("[RAW REQUEST] %s\n", string(bodyBytes))

    // 鉴权处理
    authHeader := c.GetHeader("Authorization")
    if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
        return
    }
    apiKey := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

    // 解析基础参数
    var baseReq ChatRequest
    if err := c.ShouldBindJSON(&baseReq); err != nil {
        fmt.Printf("[BIND ERROR] %v\n", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
        return
    }

    // 解析所有参数到map
    var allParams map[string]interface{}
    if err := json.Unmarshal(bodyBytes, &allParams); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format: " + err.Error()})
        return
    }

    // 提取额外参数
    extraParams := make(map[string]interface{})
    for k, v := range allParams {
        switch k {
        case "model", "messages", "stream":
            continue
        default:
            extraParams[k] = v
        }
    }

    fmt.Printf("[PARAMS] Model=%s Stream=%v Messages=%d\n",
        baseReq.Model, baseReq.Stream, len(baseReq.Messages))

    // 调用后端服务
    result, err := fangzou(apiKey, baseReq.Model, baseReq.Messages, baseReq.Stream, extraParams)
    if err != nil {
        fmt.Printf("[SERVICE ERROR] %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Service unavailable: " + err.Error()})
        return
    }

    // 流式响应处理
    if baseReq.Stream {
        stream, ok := result.(io.ReadCloser)
        if !ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Stream type assertion failed"})
            return
        }
        defer stream.Close()

        c.Stream(func(w io.Writer) bool {
            reader := bufio.NewReader(stream)
            for {
                line, err := reader.ReadBytes('\n')
                switch {
                case err == io.EOF:
                    return false
                case err != nil:
                    if !strings.Contains(err.Error(), "closed") {
                        fmt.Printf("[STREAM ERROR] %v\n", err)
                    }
                    return false
                }

                if len(bytes.TrimSpace(line)) == 0 {
                    continue
                }

                if _, err := w.Write(line); err != nil {
                    if !strings.Contains(err.Error(), "broken pipe") {
                        fmt.Printf("[WRITE ERROR] %v\n", err)
                    }
                    return false
                }
                c.Writer.Flush()
            }
        })
        return
    }

    // 普通响应处理
    c.JSON(http.StatusOK, result)
}

func main() {
    router := gin.Default()

    // 安全配置
    router.SetTrustedProxies([]string{"127.0.0.1"})
    gin.DisableConsoleColor()

    // 自定义日志中间件
    router.Use(func(c *gin.Context) {
        start := time.Now()
        c.Next()

        latency := time.Since(start)
        if c.Writer.Status() >= 500 {
            fmt.Printf("[ERROR] %s %s | %d | %v\n",
                c.Request.Method,
                c.Request.URL.Path,
                c.Writer.Status(),
                latency,
            )
        }
    })

    router.POST("/v1/chat/completions", handleRequest)

    fmt.Println("🚀 Server started on :5000")
    if err := router.Run(":5000"); err != nil {
        panic(fmt.Sprintf("Failed to start server: %v", err))
    }
}
