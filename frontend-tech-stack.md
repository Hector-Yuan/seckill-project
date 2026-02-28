# 秒杀系统前端技术栈详解文档

本文档详细解析了本项目前端部分的技术选型、架构设计及核心实现细节。

## 1. 核心设计理念

本项目前端采用 **原生技术栈 (Vanilla Stack)** 开发，即纯 HTML + CSS + JavaScript，**不依赖任何第三方框架**（如 Vue, React, jQuery, Bootstrap 等）。

*   **轻量级**：无需构建工具（Webpack/Vite），无需 `npm install`，浏览器直接运行。
*   **零依赖**：所有代码均为手写，体积极小，加载速度极快。
*   **原理性**：直观展示了前端如何与后端 RESTful API 进行交互，适合理解 HTTP 请求、JWT 鉴权和 DOM 操作的底层逻辑。

---

## 2. 技术栈详细拆解

### 2.1 HTML5 (结构层)
文件：[index.html](file:///e:/GoProjects/seckill-project/web/index.html)

使用语义化标签构建页面结构，提升代码可读性。

*   **语义化标签**：使用了 `<header>`, `<main>`, `<section>`, `<footer>` 等标签划分页面区域。
*   **表单控件**：使用原生 `<form>` 结合 `<input>` (text, password, number) 和 `<button>`。
*   **数据展示**：使用 `<table>`, `<thead>`, `<tbody>` 展示商品列表，使用 `<pre>` 标签展示原始日志信息。
*   **视口设置**：`<meta name="viewport" ...>` 确保移动端适配。

### 2.2 CSS3 (表现层)
文件：[style.css](file:///e:/GoProjects/seckill-project/web/style.css)

采用现代 CSS 特性实现响应式布局和深色主题 UI。

*   **CSS 变量 (Custom Properties)**：
    在 `:root` 中定义全局主题色，方便统一管理和修改（如 `--primary`, `--bg`, `--card`）。
    ```css
    :root {
      --bg: #0b0f19;       /* 深色背景 */
      --primary: #4f7cff;  /* 主色调 */
      /* ... */
    }
    ```
*   **布局系统 (Flexbox & Grid)**：
    *   **Flexbox**：用于导航栏、工具栏的对齐 (`justify-content: space-between`, `align-items: center`)。
    *   **CSS Grid**：用于页面主体布局 (`display: grid`)，实现了自适应的栅格系统。
    ```css
    .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
    ```
*   **响应式设计 (Media Queries)**：
    使用 `@media (max-width: 860px)` 针对小屏幕设备调整布局，将双栏布局自动切换为单栏。
*   **视觉效果**：
    *   **背景渐变**：使用了复杂的 `radial-gradient` 叠加，营造现代科技感的深色背景。
    *   **交互反馈**：按钮的 `:hover`, `:active` 状态及 `transition` 过渡动画。

### 2.3 JavaScript ES6+ (行为层)
文件：[app.js](file:///e:/GoProjects/seckill-project/web/app.js)

使用了现代 JavaScript (ES6+) 语法进行逻辑控制。

*   **常量与变量**：全面使用 `const` 和 `let`，摒弃 `var`。
*   **箭头函数**：简洁的函数表达，如 `const pad = (n) => ...`。
*   **模板字符串**：使用反引号 \`\` 进行字符串拼接和 HTML 片段生成。
*   **DOM 操作**：
    *   `document.getElementById` (封装为 `$`) 获取元素。
    *   `innerHTML` 动态渲染商品列表表格。
    *   `addEventListener` 绑定事件（点击、表单提交）。
*   **Fetch API**：
    替代 XMLHttpRequest，用于发送异步 HTTP 请求。
    *   **POST 请求**：用于注册、登录、创建商品、下单。
    *   **GET 请求**：用于获取商品列表。
    *   **Headers 处理**：统一设置 `Content-Type: application/json` 和 `Authorization` (JWT)。
*   **状态管理**：
    使用简单的全局对象 `state` 存储当前应用状态（Token、用户信息）。
    ```javascript
    const state = { token: "", user: null };
    ```
*   **本地存储 (LocalStorage)**：
    使用 `localStorage` 持久化存储 JWT Token，确保刷新页面后登录状态不丢失。

---

## 3. 核心功能实现逻辑

### 3.1 用户认证 (JWT)
1.  **登录**：用户输入账号密码 -> 发送 POST `/login` -> 后端返回 Token -> JS 存入 `state` 和 `localStorage`。
2.  **鉴权**：后续所有请求（如创建商品、下单）在 `headers` 中自动携带 `Authorization: Bearer <token>`。
3.  **注销**：清除 `state` 和 `localStorage` 中的 Token，重置 UI。

### 3.2 秒杀下单流程
1.  **列表加载**：调用 GET `/product` 获取商品数组 -> 遍历数组生成 HTML 表格行 (`<tr>`)。
2.  **点击抢购**：
    *   利用 **事件委托**，监听表格父容器的点击事件。
    *   获取按钮上的 `data-id` 属性（商品ID）。
    *   发送 POST `/order` 请求，Payload 为 `{"product_id": id}`。
3.  **结果处理**：
    *   **成功**：后端返回 200 及订单 ID，前端显示成功日志。
    *   **失败**：后端返回 400 (库存不足/重复购买) 或 500，前端捕获错误并显示原因。

### 3.3 日志系统
自定义了一个简单的日志系统，将所有操作结果实时追加到页面上的 `<pre id="log">` 区域，并在每一行前加上精确的时间戳，方便观察并发测试时的请求顺序和结果。

---

## 4. 与后端接口的对接

| 功能 | 方法 | 路径 | Payload (JSON) | Header |
| :--- | :--- | :--- | :--- | :--- |
| **注册** | POST | `/register` | `{username, password, email}` | - |
| **登录** | POST | `/login` | `{username, password}` | - |
| **商品列表** | GET | `/product` | - | Authorization |
| **创建商品** | POST | `/product` | `{name, stock}` | Authorization |
| **秒杀下单** | POST | `/order` | `{product_id}` | Authorization |

---

## 5. 总结

该前端项目虽然代码量少（约 400 行），但涵盖了现代 Web 开发的核心要素：
1.  **结构/样式/逻辑分离**。
2.  **异步编程与 API 交互**。
3.  **基于 Token 的身份验证**。
4.  **响应式布局与用户体验**。

它是一个完美的“实验台”，用于直观地测试后端秒杀系统的并发处理能力、分布式锁的有效性以及事务的一致性。
