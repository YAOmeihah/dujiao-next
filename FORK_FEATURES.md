# Dujiao-Next 二开功能说明

本文档是本 fork 的二开功能事实来源，用于功能开发、问题归属判断、上游同步、代码审查和回归测试。

- 当前官方同步基线：`v1.4.2`
- 当前 fork 发布版本：`v1.4.2-fork.3`
- 最后核对日期：`2026-08-07`
- 维护原则：只记录维护者已经确认的稳定二开，不记录设想、临时方案或未确认需求。

判断某项代码是否属于二开时，必须同时参考本文档、当前官方基线和 Git 提交历史。不能仅因文件与上游不同，就把整个文件或其中所有功能都认定为二开。

## 功能总览

| ID | 二开功能 | 状态 | 涉及端 |
| --- | --- | --- | --- |
| F01 | 游客手机号、订单密码与可选邮箱 | 保留，已适配新架构 | API、管理端、用户端 |
| F02 | 收货地址与最近地址回忆 | 保留，已适配新架构 | API、管理端、用户端 |
| F03 | VPay 支付 | 保留，已适配新架构 | API、管理端、用户端支付链路 |
| F04 | CAP 验证器 | 保留，已接入统一 captcha 模块 | API、管理端、用户端 |
| F05 | classic/default/vault 模板与二开样式兼容 | 保留 | API 设置、管理端、用户端 |
| F06 | 移动端分步结算 | 保留，使用薄适配层 | 用户端，复用 API 结算链路 |
| F07 | Fork 后台一键升级 | 保留，适配 fork 发布源 | API、管理端、GitHub Release、生产部署 |

## F01 游客手机号、订单密码与可选邮箱

### 功能说明

- 游客下单必须填写手机号和订单密码，邮箱为可选项。
- 游客查单、订单详情和后续支付访问使用下单手机号与订单密码校验身份。
- 需要收货地址时，默认可将收件手机号同步为游客查单手机号；用户手动修改后不应被无条件覆盖。
- 游客凭证通过 `Authorization: Guest ...` 请求头传递，不把订单密码放入查询字符串。
- 用户端只在当前会话保存游客查单凭证；不得把明文订单密码持久化到长期本地存储。

### 主要代码入口

- API：`internal/modules/order`、`internal/platform/http/ginutil/guest_auth.go`
- 用户端：`frontend/user/src/composables/useCheckout.ts`、`useGuestOrders.ts`、`useGuestOrderDetail.ts`
- 凭证工具：`frontend/user/src/utils/guestOrderAuth.ts`
- 页面：`frontend/user/src/views/GuestOrders.vue`、`GuestOrderDetail.vue`、`Checkout.vue`

### 数据与接口约束

- 游客订单包含 `guest_phone`、可选 `guest_email` 和订单密码凭证。
- 创建订单和订单预览必须保持相同的手机号、密码、邮箱校验规则。
- 管理端订单查询和详情需要继续展示或支持筛选游客手机号、邮箱。
- 不兼容迁移前仅使用邮箱与订单密码查询的历史游客订单；新架构以手机号与订单密码作为游客身份凭证。

### 回归检查

- 游客下单缺少手机号或订单密码时必须被拒绝；邮箱为空时允许下单。
- 手机号格式错误、邮箱填写后格式错误时应给出对应错误。
- 正确手机号和密码可查单、查看详情和继续支付；错误凭证不可访问订单。
- 收货手机号自动同步、手动修改和再次编辑的行为正确。
- 网络请求 URL、日志和错误信息中不得出现订单密码。

### 上游同步注意

- 上游若仍以邮箱作为游客主身份，不得直接覆盖本 fork 的手机号鉴权。
- 合并订单 DTO、handler、presenter 或查询逻辑时，要同时检查预览、创建、查单、详情和支付链路，避免只适配其中一个入口。

## F02 收货地址与最近地址回忆

### 功能说明

- 商品可通过 `requires_shipping_address` 指定订单必须填写收货地址。
- 地址包含收件人、手机号、省、市、区/县、街道/乡镇和详细地址。
- 省、市、区/县、街道/乡镇使用区划代码逐级选择；API 会校验层级关系并以可信区划名称归一化。
- 游客在需要收货地址时可保存并复用最近一次地址，也可重新填写、清空表单或删除本地记录。
- 最近地址只保存在浏览器本地，不创建服务端地址簿。
- 浏览器本地存储不可用、读写失败或记录损坏时，不得阻断地址填写和订单提交。

### 主要代码入口

- 区划 API：`internal/modules/addressdivision`
- 订单校验与存储：`internal/modules/order/application/value.go`、`internal/modules/order`
- 商品开关：`internal/modules/catalog/product`、`frontend/admin/src/views/admin/components/ProductEditModal.vue`
- 用户端：`frontend/user/src/components/checkout/RegionSelector.vue`、`GuestShippingAddressRecallCard.vue`
- 本地回忆：`frontend/user/src/composables/useGuestShippingAddressRecall.ts`、`useCheckout.ts`

### 数据与接口约束

- 收货地址保存为订单快照，不能依赖商品或用户资料后续变化。
- API 必须校验 `province_code`、`city_code`、`district_code`、`township_code` 的层级归属。
- 只有订单中存在要求收货地址的商品时，预览和创建订单才强制要求地址。
- 地址订单前端要求收件人姓名至少 2 个字符、详细地址至少 4 个字符；收货手机号在提交前统一规范化为可选 `+` 和 6–20 位数字。

### 回归检查

- 管理端可开启和关闭商品的收货地址要求，用户端商品与结算状态同步变化。
- 四级区划选择、返回上一级重选、详细地址和手机号校验正确。
- 缺字段、伪造名称、错误区划代码或断裂层级必须被 API 拒绝。
- 游客最近地址可使用、重填、清空和删除；会员结算或无收货要求时不应错误显示。
- 订单创建后，用户端和管理端订单详情能读取完整地址快照。

### 上游同步注意

- 合并商品模型或订单校验时不能丢失 `requires_shipping_address` 和 `shipping_address`。
- 上游若新增用户地址簿，应先分析与本地“游客最近地址”是否重复，再由维护者决定合并方式，不能直接删除本地回忆功能。

## F03 VPay 支付

### 功能说明

- 管理端可创建 `vpay` 类型支付渠道并配置 VPay 网关。
- 支持创建 VPay 订单、跳转支付页、同步返回和异步回调。
- VPay 通过上游统一 `GatewayProvider` 注册与支付应用服务接入，不维护平行支付主流程。
- 当前仅支持 `redirect` 交互模式。
- 当前仅支持 `CNY`；系统不为 VPay 进行汇率换算，非 `CNY` 支付请求在调用网关前拒绝。
- 支持 `MD5` 和 `HMAC_SHA256` 签名，并验证回调签名、订单金额和实付金额。

### 主要代码入口

- VPay 协议：`internal/modules/payment/infrastructure/gateway/vpay`
- 统一网关适配：`internal/modules/payment/infrastructure/gateway/adapters/vpay`
- 回调：`internal/modules/payment/transport/http/callback/vpay_callback.go`
- 支付应用服务：`internal/modules/payment/application/payment_service_provider.go`
- 管理端配置：`frontend/admin/src/views/admin/components/PaymentChannelModal.vue`

### 配置约束

- 必填核心配置：`gateway_url`、签名密钥、`notify_url`、`return_url`。
- 可配置 `sign_type`、创建订单路径和支付页路径；默认路径必须继续兼容 VPay 标准部署。
- `key`、`secret_key`、`sign_key` 等密钥属于敏感配置，接口和日志不得回显完整值。

### 回归检查

- 管理端可保存、回显脱敏配置并正确校验缺失或非法字段。
- 创建支付时金额、订单号、渠道类型、回调地址和签名正确，最终返回可访问的跳转地址。
- 非 `redirect` 模式和不支持的渠道类型必须被拒绝。
- 合法回调可更新订单；错误签名、错误金额或错误支付渠道不能更新订单。
- 重复回调保持幂等，支付完成后的用户端跳转和订单状态刷新正常。

### 上游同步注意

- 上游支付 provider 注册、回调路由或支付状态机变化时，应更新 VPay 适配器，不复制或恢复旧支付主流程。
- 合并通用回调逻辑时必须保留 VPay 对 `reallyPrice` 和签名类型的验证。

## F04 CAP 验证器

### 功能说明

- 验证器展示名称固定为“CAP 验证器”，provider 值为 `cap`。
- CAP 已接入统一 captcha 模块，与图片验证码和 Turnstile 共用配置、公开配置、场景开关和验证接口。
- 支持后台登录、用户登录、注册验证码、重置密码验证码、游客下单和礼品卡兑换等统一 captcha 场景。
- 不保留旧版独立 CAP 服务层或平行验证码路由。

### 主要代码入口

- 配置：`config.yml.example`、`internal/config/config.go`
- captcha 服务：`internal/modules/captcha`
- CAP 客户端：`internal/modules/captcha/infrastructure/cap`
- 动态设置：`internal/modules/settings/schema/security/captcha.go`
- 管理端：`frontend/admin/src/components/captcha/CapCaptcha.vue`、`SettingsCaptchaTab.vue`
- 用户端：`frontend/user/src/components/captcha/CapCaptcha.vue` 及各认证、结算场景 composable

### 配置约束

- CAP 配置包含 `endpoint`、`site_key`、`secret_key` 和 `timeout_ms`。
- 公开站点配置只能返回前端渲染所需数据，不能泄露 `secret_key`。
- CAP endpoint、site key 或 secret key 不完整时，启用相关场景必须得到明确配置错误，不能静默绕过验证码。

### 回归检查

- 管理端可选择 CAP、保存设置并正确处理“保留已有 secret”语义。
- 用户端和管理端组件可加载、产出 token、失败后重置并重新验证。
- 各场景开关分别生效；未启用的场景不应强制 CAP。
- 空 token、无效 token、超时和 CAP 服务不可用时返回一致错误。
- 公开配置、API 响应和日志中不包含 CAP secret。

### 上游同步注意

- 上游调整 captcha contract、payload 或公开设置时，CAP 必须继续作为统一 provider 适配，不能重新散落到各 handler。
- 新增 captcha 场景时先评估 CAP 是否应同时支持，并补最小回归测试。

## F05 classic/default/vault 模板与二开样式兼容

### 功能说明

- `classic` 是默认模板；配置层的 `default` 语义继续映射到 classic。
- 保留 `vault` 模板及其独立页面、设计令牌和样式作用域。
- vault 缺少某个页面时回退到 classic 页面，避免模板开发不完整导致路由不可用。
- classic 与 vault 结算页共享同一套移动端分步结算组件，各页面只初始化一个 `useCheckout` 状态实例。
- 二开表单、输入框、按钮、支付方式、弹窗、地址选择和结算控件统一复用现有主题变量与 `theme-*` 工具类。
- Teleport 到 `body` 的弹窗和浮层必须获得当前模板的背景、边框和文字令牌，不能变成透明背景。

### 主要代码入口

- 模板注册与回退：`frontend/user/src/templates/registry.ts`
- vault 页面和样式：`frontend/user/src/templates/vault`
- classic/default 公共样式：`frontend/user/src/style.css`
- 主题状态：`frontend/user/src/utils/theme.ts`
- 模板常量与站点设置：`internal/constants/constants.go`、站点设置相关模块

### 回归检查

- 管理端切换模板后，classic/default 与 vault 路由均可打开。
- vault 缺页时能回退 classic，不出现空白页或循环加载。
- 结算输入框、购买方式、支付渠道、CAP、公告、地址选择弹窗在亮色和暗色下均有正确背景与边框。
- 桌面和移动端不存在透明弹窗、不可见文字、内容遮挡或底部安全区缺失。
- 新增二开界面优先复用现有组件和主题类；现有样式不能满足时才新增局部样式。

### 上游同步注意

- 同步 `style.css`、模板 registry、Dialog/Popover 组件或结算页面时，重点检查上游删除样式类、CSS 变量改名和 Teleport 作用域变化。
- 不因上游重做某个页面而整体覆盖 vault；按页面判断是复用、适配还是保留 fork 实现。

## F06 移动端分步结算

### 功能说明

- 移动端结算按商品、收货、购买信息、优惠券和支付方式组织为分步流程。
- 不需要收货地址时自动跳过收货步骤。
- reseller 站点不展示优惠券步骤；手工必填文本和游客订单密码只要求非空，不增加后端不存在的最小长度限制。
- 需要收货地址时，移动端继续执行收件人姓名至少 2 个字符、详细地址至少 4 个字符及手机号格式校验。
- 收货、购买信息和支付方式需要确认；确认后修改数据会使步骤重新变为待确认。
- 校验失败时展开对应步骤、展示具体错误并定位到相关控件。
- 底部固定操作栏根据当前步骤提供保存、继续、选择支付或提交订单操作，并适配移动设备安全区。
- 移动端只维护界面步骤状态；订单预览、库存、地址、游客信息、CAP、钱包、支付渠道和提交逻辑必须继续复用 `useCheckout`。

### 主要代码入口

- 页面接入：`frontend/user/src/views/Checkout.vue`
- 移动组件：`frontend/user/src/components/checkout/mobile`
- 纯流程规则：`frontend/user/src/composables/useMobileCheckoutFlow.ts`
- 薄适配层：`frontend/user/src/composables/useMobileCheckoutAdapter.ts`
- 结算业务状态：`frontend/user/src/composables/useCheckout.ts`

### 回归检查

- 购物车结算和立即购买均能正确显示商品与步骤。
- 有/无收货地址、会员/游客、有/无手工表单、有/无优惠券、余额/在线支付组合均能推进。
- 已确认步骤修改后必须重新确认；未完成步骤不能直接提交。
- CAP、手机号、邮箱、地址、支付金额和支付渠道错误会定位到正确步骤。
- 提交过程中禁用重复操作，不产生重复订单。
- 固定底栏不遮挡内容，横竖屏和常见移动宽度下无错位。

### 上游同步注意

- 上游修改 `useCheckout` 返回值或结算页面结构时，只更新 adapter 和组件绑定，避免把业务逻辑复制进移动流程。
- 同步后重点检查 desktop 结算仍使用同一业务状态，移动适配不能反向改变桌面行为。

## F07 Fork 后台一键升级

### 功能说明

- 复用官方后台一键升级、自更新校验、二进制原子替换和回滚机制，不维护平行更新流程。
- 更新源固定为本 fork 仓库 `YAOmeihah/dujiao-next`，禁止从官方 `dujiao-next/dujiao-next` 下载二进制，避免升级后丢失 F01-F07 二开功能。
- 自动更新只接受严格符合 `vX.Y.Z-fork.N` 的正式 GitHub Release；同一上游版本递增 `fork.N`，升级上游基线后从 `fork.1` 重新开始。
- GitHub Actions 使用 GoReleaser 构建内嵌 API、管理端和用户端的单一二进制；后台升级会按当前平台下载 fork Release、校验 checksums 后替换正在运行的可执行文件。
- 裸二进制生产环境使用稳定路径 `/www/wwwroot/djxy/api/dujiao-next`，Supervisor 不得指向 `releases/<版本>/dujiao-next`，也不得用指向版本目录的软链接代替稳定文件。
- 宝塔 Supervisor 不会被官方更新器识别为可自动重启的 systemd。二进制替换完成后必须在宝塔手动重启进程，并检查健康状态和前后台版本。

### 更新边界与生产目录

- 后台一键升级只提取并替换 `dujiao-next` 二进制，同时维护同目录的 `.backup`、锁和升级元数据；不会更新 `config.yml`、`config.yml.example`、数据库、`uploads/`、`logs/` 或 `data/address_divisions/`。
- 上游或 fork 新版本增加配置项时，必须检查 Release notes 和新版 `config.yml.example`，按需人工合并到生产 `config.yml`，不能假设一键升级会补齐配置。
- F02 依赖的 `data/address_divisions/` 必须长期保留。若后续 Release 更新区划数据，需要单独安排人工更新和验证。
- 生产目录迁移遗留物只在备份、停服检查和升级验证完成后人工清理；不得在应用、自更新器或部署脚本中加入自动清理生产目录的逻辑。
- 稳定运行所需内容包括 `dujiao-next`、`config.yml`、`data/address_divisions/`、`uploads/`、`logs/`、当前数据库所需文件，以及自更新器生成的备份和元数据文件。

### 主要代码入口

- fork 发布源和标签校验：`internal/version/release.go`
- 官方自更新机制：`internal/selfupdate`
- 后台更新接口：`internal/platform/http/system/update_handler.go`
- 管理端更新界面：`frontend/admin/src/components/SystemUpdateDialog.vue`
- fork Release 构建：`.github/workflows/release.yml`、`.goreleaser.yaml`

### 回归检查

- 后台检查更新返回的来源必须是 `https://github.com/YAOmeihah/dujiao-next/releases`，不得回退到官方仓库。
- `vX.Y.Z-fork.N` 可参与版本比较和升级；普通官方 tag、RC tag、缺少 `v` 或 `fork.N` 非法的 tag 必须被拒绝。
- fork Release 必须包含当前平台归档和 checksums，归档内二进制必须同时包含 API、管理端和用户端。
- 升级前后 `config.yml`、数据库、上传文件和区划数据保持不变；旧二进制备份及更新元数据可用于官方回滚流程。
- 宝塔环境升级完成后不会错误宣称已自动重启；人工重启后 `/health`、商城首页和隐藏后台路径均可访问，运行版本与目标 fork tag 一致。

### 上游同步注意

- 每次上游修改 `internal/version`、`internal/selfupdate`、后台更新接口、管理端更新界面、`.goreleaser.yaml` 或 Release workflow 时，都必须同步官方安全修复和功能变化，同时重新核对 fork 发布源与标签约束。
- 解决冲突后必须确认 `repoOwner` 仍为 `YAOmeihah`，Release tag 仍只接受 `vX.Y.Z-fork.N`，不能因采用上游文件而恢复为 `dujiao-next/dujiao-next`。
- 上游若调整附件命名、校验文件、压缩格式、二进制替换、回滚元数据或进程重启机制，必须同步更新 fork 适配和回归测试，确保 GitHub Release 产物仍能被后台更新器识别。
- 上游若开始自动修改配置文件或附带数据，必须先评估对生产密钥、数据库、上传文件和 F02 区划数据的影响，再决定是否跟进；不得在冲突处理中直接启用。

## 已由上游接管

### 首页公告弹窗

- 首页公告弹窗最初属于本 fork 需求，但官方现已提供 `home_announcement` 设置和展示链路。
- 当前以官方实现为准，本 fork 不再维护旧 `is_home_popup` 字段、接口或前端分支。
- 同步上游时必须保留官方公告模型和展示逻辑；发现旧实现残留时应删除残留，而不是恢复旧路径。
- 二开模板仍需保证官方公告弹窗在 classic/default/vault 下样式可见；这属于 F05 的兼容责任，不代表重新维护公告业务逻辑。

## 明确不属于本项目二开

- SKU 批发价、会员价及相关后台页面。
- 分销、reseller、站点定制和多站点能力。
- 上述功能按官方实现处理。只有当本 fork 的六项二开直接破坏这些官方功能时，才作为二开兼容问题处理。

## 上游同步与问题归属

每次同步上游后按以下顺序判断问题：

1. 确认官方同步基线和 fork 合并前后的提交范围。
2. 先查本文档，确认相关行为是否属于 F01-F06、上游已接管或明确非二开。
3. 对比官方同版本代码，判断问题在官方基线中是否已经存在。
4. 只修复二开自身错误，以及二开与上游合并、迁移、接口变化造成的兼容错误。
5. 上游原本存在且未被二开改动触发的问题只记录结论，不在 fork 中修复。
6. 二开与上游新功能在产品行为上冲突时，列出两方行为、影响和后续同步成本，由维护者决定保留哪一方。

## 文档维护规则

- 新增功能只有在维护者明确确认“属于二开”后，才能分配新的 `Fxx` 编号并写入本文档。
- 修改二开行为时，在同一开发周期更新功能说明、代码入口、回归检查和同步注意事项。
- 删除二开或改由上游接管时，不直接抹掉历史；移动到“已由上游接管”并说明禁止恢复的旧路径。
- 代码入口只记录稳定模块和关键文件，不罗列所有文件，避免普通重构造成无意义文档变更。
- 上游基线或 fork 基础版本升级后，更新文档头部版本并新增一条变更记录。
- `AGENTS.md` 只维护工程规则和本文档入口，不复制详细功能说明，避免两个事实来源不一致。

## 变更记录

| 日期 | 官方基线 | 变更 |
| --- | --- | --- |
| 2026-08-07 | `v1.4.2` | 发布 `v1.4.2-fork.3`；新增 F07：官方后台一键升级适配 fork Release，固定更新源、fork tag 规则、生产二进制路径和上游同步核对要求。 |
| 2026-08-07 | `v1.4.2` | 明确历史游客订单不做邮箱兼容、VPay 仅支持 CNY；补充地址存储容错、手机号规范化、移动校验边界和 vault 移动结算共享实现。 |
| 2026-08-06 | `v1.4.2` | 在官方单仓新架构上确认并记录 F01-F06；首页公告改由官方实现接管；明确 SKU 批发价/会员价、分销和站点定制不属于二开。 |
