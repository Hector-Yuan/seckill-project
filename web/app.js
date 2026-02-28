const state = {
  token: localStorage.getItem("seckill_token") || "",
  user: JSON.parse(localStorage.getItem("seckill_user") || "null"),
};

function $(id) {
  return document.getElementById(id);
}

function nowStr() {
  const d = new Date();
  const pad = (n) => String(n).padStart(2, "0");
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

function log(message, payload) {
  const el = $("log");
  if (!el) return;
  const line =
    payload === undefined ? `[${nowStr()}] ${message}\n` : `[${nowStr()}] ${message}\n${JSON.stringify(payload, null, 2)}\n`;
  el.textContent = line + el.textContent;
}

function saveSession(token, user) {
  state.token = token || "";
  state.user = user || null;
  localStorage.setItem("seckill_token", state.token);
  localStorage.setItem("seckill_user", JSON.stringify(state.user));
}

function clearSession() {
  state.token = "";
  state.user = null;
  localStorage.removeItem("seckill_token");
  localStorage.removeItem("seckill_user");
}

// 强制登录检查
function requireAuth(requiredRole) {
  if (!state.token) {
    alert("请先登录！");
    window.location.href = "./login.html";
    return;
  }

  if (requiredRole && (!state.user || state.user.role !== requiredRole)) {
    alert("权限不足！当前页面需要 " + requiredRole + " 权限");
    window.location.href = "./index.html";
    return;
  }

  // 如果有显示用户名的元素，更新它
  const userEl = document.getElementById("currentUsername");
  if (userEl && state.user) {
    const roleText = state.user.role === "admin" ? " (管理员)" : "";
    userEl.textContent = state.user.username + roleText;
  }
}

// 登出
function logout() {
  clearSession();
  log("已退出登录");
}

async function request(path, options = {}) {
  const headers = { "Content-Type": "application/json" };
  if (state.token) {
    headers["Authorization"] = "Bearer " + state.token;
  }
  const res = await fetch(path, {
    ...options,
    headers: { ...headers, ...options.headers },
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || data.err || "未知错误");
  }
  return data;
}

// 业务逻辑函数

async function handleRegister(data) {
  try {
    await request("/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
    log("注册成功", data);
    alert("注册成功，请去登录");
    window.location.href = "./login.html";
  } catch (err) {
    log("注册失败: " + err.message);
    alert("注册失败: " + err.message);
  }
}

async function handleLogin(data) {
  try {
    const res = await request("/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
    log("登录成功", res);
    saveSession(res.token, { username: res.username, role: res.role });
    alert("登录成功");
    
    if (res.role === "admin") {
      window.location.href = "./admin.html";
    } else {
      window.location.href = "./list.html";
    }
  } catch (err) {
    log("登录失败: " + err.message);
    alert("登录失败: " + err.message);
  }
}

async function handleCreateProduct(data) {
  try {
    const res = await request("/product", {
      method: "POST",
      body: JSON.stringify(data),
    });
    log("创建商品成功", res);
    alert("创建商品成功");
    // 如果有刷新逻辑可以调用，但这里 admin 页面没有列表，不需要刷新
  } catch (err) {
    log("创建商品失败: " + err.message);
    alert("创建商品失败: " + err.message);
  }
}

async function handleLoadProducts() {
  try {
    const res = await request("/product");
    const tbody = $("productTbody");
    if (!tbody) return;
    
    tbody.innerHTML = "";
    const list = res.data || [];
    log(`加载商品列表成功: ${list.length} 个商品`);

    if (list.length === 0) {
      tbody.innerHTML = `<tr><td colspan="4" style="text-align: center">暂无商品</td></tr>`;
      return;
    }

    list.forEach((p) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td>${p.ID}</td>
        <td>${p.name}</td>
        <td>${p.stock}</td>
        <td>
           <button class="btn primary sm" data-action="order" data-id="${p.ID}">抢购</button>
        </td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {
    log("加载商品失败: " + err.message);
    const tbody = $("productTbody");
    if (tbody) tbody.innerHTML = `<tr><td colspan="4" style="color: var(--danger)">加载失败: ${err.message}</td></tr>`;
  }
}

async function handleOrder(productId) {
  const id = parseInt(productId, 10);
  try {
    log(`正在抢购商品 ID: ${id}...`);
    const res = await request("/order", {
      method: "POST",
      body: JSON.stringify({ product_id: id }),
    });
    log("抢购成功！", res);
    alert("抢购成功！订单ID: " + res.order_id);
    handleLoadProducts(); // 刷新库存
  } catch (err) {
    log("抢购失败: " + err.message);
    alert("抢购失败: " + err.message);
  }
}
