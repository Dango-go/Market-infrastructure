import { startTransition, useDeferredValue, useEffect, useMemo, useState, type CSSProperties } from "react";
import {
  BadgeCheck,
  Boxes,
  ChevronRight,
  CircuitBoard,
  Clock3,
  Cpu,
  Heart,
  LoaderCircle,
  LogOut,
  MapPin,
  Minus,
  PackageCheck,
  Plus,
  Search,
  ShieldCheck,
  ShoppingBag,
  SlidersHorizontal,
  Sparkles,
  Star,
  Truck,
  UserRound,
  WalletCards,
  Zap
} from "lucide-react";
import {
  addCartItem,
  addWishlistItem,
  checkoutCart,
  confirmPayment,
  createOrder,
  createPayment,
  createShipment,
  getAccessToken,
  getActiveCart,
  getCurrentAccount,
  getReviewSummary,
  listCatalogProducts,
  listCategories,
  listOrders,
  listPayments,
  listShipments,
  listTrendingRecommendations,
  listWishlist,
  loginAccount,
  logoutSession,
  refreshSession,
  registerAccount,
  removeWishlistItem,
  setAccessToken,
  trackPageView,
  updateCartItem,
  updateOrderStatus,
  type Account,
  type AuthResponse,
  type Cart,
  type CatalogCategory,
  type CatalogProduct,
  type Order,
  type Payment,
  type Recommendation,
  type ReviewSummary,
  type Shipment,
  type WishlistItem
} from "./api";

type Product = {
  id: string;
  name: string;
  category: string;
  tag: string;
  price: number;
  oldPrice?: number;
  rating: number;
  reviews: number;
  stock: string;
  accent: string;
  imagePosition: string;
  delivery: string;
  specs: string[];
  imageUrl: string;
};

type CategoryOption = {
  label: string;
  value: string;
};

type StorefrontState = {
  products: Product[];
  categories: CategoryOption[];
  recommendations: Product[];
  averageRating: number;
  totalReviews: number;
  live: boolean;
  note: string;
};

type AuthMode = "login" | "register";

type OrderBundle = {
  orders: Order[];
  payments: Payment[];
  shipments: Shipment[];
};

const fallbackProducts: Product[] = [
  {
    id: "esp32-s3-lab-kit",
    name: "ESP32-S3 Vision Lab Kit",
    category: "Dev boards",
    tag: "AI camera",
    price: 42,
    oldPrice: 58,
    rating: 4.9,
    reviews: 128,
    stock: "186 in Kyiv hub",
    accent: "lime",
    imagePosition: "62% 44%",
    delivery: "Today, 18:00",
    specs: ["Wi-Fi 6", "8MB PSRAM", "USB-C"],
    imageUrl: "/images/embedded-hero.png"
  },
  {
    id: "pi-carrier-cm4",
    name: "CM4 Industrial Carrier",
    category: "Single-board",
    tag: "DIN rail",
    price: 89,
    rating: 4.8,
    reviews: 92,
    stock: "44 in stock",
    accent: "cyan",
    imagePosition: "72% 52%",
    delivery: "Tomorrow",
    specs: ["M.2 slot", "PoE ready", "RS485"],
    imageUrl: "/images/embedded-hero.png"
  },
  {
    id: "sensor-grid-pack",
    name: "Sensor Grid Pack",
    category: "Sensors",
    tag: "12 modules",
    price: 34,
    oldPrice: 41,
    rating: 4.7,
    reviews: 211,
    stock: "Hot batch",
    accent: "coral",
    imagePosition: "49% 56%",
    delivery: "Today, 21:00",
    specs: ["IMU", "ToF", "BME280"],
    imageUrl: "/images/embedded-hero.png"
  },
  {
    id: "oled-micro-stack",
    name: "OLED Micro Stack",
    category: "Displays",
    tag: "Prototype",
    price: 18,
    rating: 4.6,
    reviews: 76,
    stock: "310 available",
    accent: "mint",
    imagePosition: "43% 38%",
    delivery: "2 days",
    specs: ["0.96 inch", "I2C", "Low power"],
    imageUrl: "/images/embedded-hero.png"
  },
  {
    id: "robotics-motion-bundle",
    name: "Robotics Motion Bundle",
    category: "Actuators",
    tag: "Starter",
    price: 63,
    rating: 4.9,
    reviews: 54,
    stock: "72 in stock",
    accent: "yellow",
    imagePosition: "55% 63%",
    delivery: "Tomorrow",
    specs: ["Servos", "Drivers", "Jumper kit"],
    imageUrl: "/images/embedded-hero.png"
  },
  {
    id: "edge-power-shield",
    name: "Edge Power Shield",
    category: "Power",
    tag: "UPS",
    price: 27,
    oldPrice: 35,
    rating: 4.5,
    reviews: 38,
    stock: "Low stock",
    accent: "rose",
    imagePosition: "67% 48%",
    delivery: "Friday",
    specs: ["LiPo", "Fuel gauge", "5V boost"],
    imageUrl: "/images/embedded-hero.png"
  }
];

const fallbackCategories: CategoryOption[] = [
  { label: "All", value: "all" },
  { label: "Dev boards", value: "dev-boards" },
  { label: "Single-board", value: "single-board" },
  { label: "Sensors", value: "sensors" },
  { label: "Displays", value: "displays" },
  { label: "Actuators", value: "actuators" },
  { label: "Power", value: "power" }
];

const deliveryOptions = [
  { title: "Nova lab courier", time: "Today 18:00", price: "$4.20", icon: Truck },
  { title: "Pickup pod", time: "24/7 locker", price: "$1.10", icon: PackageCheck },
  { title: "Workshop drop", time: "2 hour slot", price: "$7.50", icon: MapPin }
];

const accentCycle = ["lime", "cyan", "coral", "mint", "yellow", "rose"];
const imagePositions = ["62% 44%", "72% 52%", "49% 56%", "43% 38%", "55% 63%", "67% 48%"];
const deliveryCycle = ["Today, 18:00", "Tomorrow", "Today, 21:00", "2 days", "Friday"];

function App() {
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("all");
  const [guestCart, setGuestCart] = useState<Record<string, number>>({ "esp32-s3-lab-kit": 1, "sensor-grid-pack": 1 });
  const [delivery, setDelivery] = useState(0);
  const [storefront, setStorefront] = useState<StorefrontState>(() => ({
    products: fallbackProducts,
    categories: fallbackCategories,
    recommendations: fallbackProducts.slice(1, 4),
    averageRating: 4.8,
    totalReviews: 599,
    live: false,
    note: "Demo shelf while the catalog is still waking up."
  }));
  const [account, setAccount] = useState<Account | null>(null);
  const [authMode, setAuthMode] = useState<AuthMode>("login");
  const [authLoading, setAuthLoading] = useState(true);
  const [authBusy, setAuthBusy] = useState(false);
  const [authError, setAuthError] = useState("");
  const [authForm, setAuthForm] = useState({ email: "", username: "", identifier: "", password: "" });
  const [cartData, setCartData] = useState<Cart | null>(null);
  const [wishlistItems, setWishlistItems] = useState<WishlistItem[]>([]);
  const [orderBundle, setOrderBundle] = useState<OrderBundle>({ orders: [], payments: [], shipments: [] });
  const [opsLoading, setOpsLoading] = useState(false);
  const [cartBusy, setCartBusy] = useState(false);
  const [checkoutBusy, setCheckoutBusy] = useState(false);
  const [checkoutNotice, setCheckoutNotice] = useState("");
  const [checkoutForm, setCheckoutForm] = useState({
    address: "Kyiv, Makerspace 14, Bench 3",
    note: "Please include anti-static packaging.",
    paymentMethod: "card"
  });
  const [loading, setLoading] = useState(true);
  const deferredQuery = useDeferredValue(query);

  useEffect(() => {
    trackPageView(window.location.pathname).catch(() => undefined);
  }, []);

  useEffect(() => {
    void bootstrapAccount();
  }, []);

  useEffect(() => {
    const controller = new AbortController();

    async function loadStorefront() {
      setLoading(true);
      try {
        const [catalogPage, categoriesPayload, trendingPayload] = await Promise.all([
          listCatalogProducts({
            query: deferredQuery.trim() || undefined,
            categorySlug: category === "all" ? undefined : category
          }),
          listCategories(),
          listTrendingRecommendations(4)
        ]);

        if (controller.signal.aborted) return;

        const categories = categoriesPayload.items.length
          ? [{ label: "All", value: "all" }, ...categoriesPayload.items.map((item) => ({ label: item.name, value: item.slug }))]
          : fallbackCategories;

        const products = catalogPage.items.length
          ? await hydrateProducts(catalogPage.items, categoriesPayload.items, controller.signal)
          : fallbackProducts;

        const recommendations = mapRecommendations(trendingPayload.items, products);
        const live = catalogPage.items.length > 0;
        const totalReviews = products.reduce((sum, product) => sum + product.reviews, 0);
        const averageRating = products.length
          ? Number((products.reduce((sum, product) => sum + product.rating, 0) / products.length).toFixed(1))
          : 4.8;

        startTransition(() => {
          setStorefront({
            products,
            categories,
            recommendations,
            averageRating,
            totalReviews,
            live,
            note: live
              ? "Live through api-gateway. Orders, payments and shipments are ready to ride on top."
              : "Gateway is wired. Demo shelf stays in place until the catalog has real seeded products."
          });
        });
      } catch {
        if (controller.signal.aborted) return;
        startTransition(() => {
          setStorefront((current) => ({
            ...current,
            products: fallbackProducts,
            categories: fallbackCategories,
            recommendations: fallbackProducts.slice(1, 4),
            averageRating: 4.8,
            totalReviews: 599,
            live: false,
            note: "Gateway is unreachable right now, so the storefront is keeping the demo shelf warm."
          }));
        });
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      }
    }

    loadStorefront();
    return () => controller.abort();
  }, [deferredQuery, category]);

  useEffect(() => {
    if (!account) {
      setCartData(null);
      setWishlistItems([]);
      setOrderBundle({ orders: [], payments: [], shipments: [] });
      return;
    }
    void loadOperationalData();
  }, [account]);

  const visibleProducts = storefront.products;
  const productByID = useMemo(() => new Map(storefront.products.map((product) => [product.id, product])), [storefront.products]);
  const effectiveCart = useMemo(() => {
    if (cartData) {
      const next: Record<string, number> = {};
      for (const item of cartData.items) next[item.product_id] = item.quantity;
      return next;
    }
    return guestCart;
  }, [cartData, guestCart]);
  const cartItems = useMemo(() => storefront.products.filter((product) => effectiveCart[product.id]), [effectiveCart, storefront.products]);
  const wishlistProductIDs = useMemo(() => new Set(wishlistItems.map((item) => item.product_id)), [wishlistItems]);
  const subtotal = cartItems.reduce((sum, product) => sum + product.price * effectiveCart[product.id], 0);
  const deliveryPrice = Number(deliveryOptions[delivery].price.replace("$", ""));
  const total = subtotal + deliveryPrice;
  const stats = [
    { icon: Cpu, value: String(storefront.products.length), label: storefront.live ? "catalog modules live" : "demo modules ready" },
    { icon: Zap, value: String(storefront.recommendations.length), label: "gateway recommendation picks" },
    { icon: ShieldCheck, value: storefront.averageRating.toFixed(1), label: `${storefront.totalReviews} review signals` },
    { icon: Boxes, value: String(orderBundle.orders.length || storefront.products.filter((item) => item.oldPrice).length), label: account ? "orders tracked" : "discounted build kits" }
  ];
  const latestOrder = orderBundle.orders[0] ?? null;
  const latestPayment = latestOrder ? orderBundle.payments.find((payment) => payment.order_id === latestOrder.id) ?? null : null;
  const latestShipment = latestOrder ? orderBundle.shipments.find((shipment) => shipment.order_id === latestOrder.id) ?? null : null;

  async function bootstrapAccount() {
    setAuthLoading(true);
    try {
      if (getAccessToken()) {
        try {
          const current = await getCurrentAccount();
          setAccount(current);
          setAuthError("");
          return;
        } catch {
          // Fall through to refresh when the access token expired.
        }
      }
      const refreshed = await refreshSession();
      applyAuthResult(refreshed);
    } catch {
      setAccessToken(null);
      setAccount(null);
    } finally {
      setAuthLoading(false);
    }
  }

  async function loadOperationalData() {
    setOpsLoading(true);
    try {
      const [cart, wishlistPage, ordersPage, paymentsPage, shipmentsPage] = await Promise.all([
        getActiveCart(),
        listWishlist(),
        listOrders(),
        listPayments(),
        listShipments()
      ]);
      setCartData(cart);
      setWishlistItems(wishlistPage.items);
      setOrderBundle({ orders: ordersPage.items, payments: paymentsPage.items, shipments: shipmentsPage.items });
    } finally {
      setOpsLoading(false);
    }
  }

  async function syncGuestCart(items: Record<string, number>) {
    const existingCart = await getActiveCart();
    const existingQuantities = new Map(existingCart.items.map((item) => [item.product_id, item.quantity]));
    const entries = Object.entries(items);

    for (const [productID, quantity] of entries) {
      const product = productByID.get(productID);
      if (!product) continue;

      const currentQuantity = existingQuantities.get(productID) ?? 0;
      const missingQuantity = Math.max(quantity - currentQuantity, 0);
      if (!missingQuantity) continue;

      const nextCart = await addCartItem({
        productID,
        quantity: missingQuantity,
        unitPriceCents: toCents(product.price)
      });
      for (const item of nextCart.items) {
        existingQuantities.set(item.product_id, item.quantity);
      }
    }

    setGuestCart({});
  }

  function applyAuthResult(result: AuthResponse) {
    setAccessToken(result.access_token);
    setAccount(result.account);
    setAuthError("");
  }

  function validateAuthForm() {
    if (authMode === "login") {
      if (!authForm.identifier.trim()) return "Enter your email or username.";
      if (!authForm.password) return "Enter your password.";
      return "";
    }

    if (!authForm.email.trim()) return "Email is required.";
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(authForm.email.trim())) return "Enter a valid email address.";
    if (authForm.username.trim().length < 3) return "Username must be at least 3 characters.";
    if (authForm.password.length < 10) return "Password must be at least 10 characters.";
    return "";
  }

  async function submitAuth() {
    const validationError = validateAuthForm();
    if (validationError) {
      setAuthError(validationError);
      return;
    }

    setAuthBusy(true);
    setAuthError("");
    try {
      const result = authMode === "login"
        ? await loginAccount({ identifier: authForm.identifier, password: authForm.password })
        : await registerAccount({ email: authForm.email, username: authForm.username, password: authForm.password });
      applyAuthResult(result);
      if (Object.keys(guestCart).length) {
        await syncGuestCart(guestCart);
      }
      await loadOperationalData();
    } catch (error) {
      setAuthError(error instanceof Error ? error.message : "Authentication failed");
    } finally {
      setAuthBusy(false);
    }
  }

  async function signOut() {
    setAuthBusy(true);
    try {
      await logoutSession();
    } catch {
      // Best effort logout.
    } finally {
      setAccessToken(null);
      setAccount(null);
      setCartData(null);
      setWishlistItems([]);
      setOrderBundle({ orders: [], payments: [], shipments: [] });
      setAuthBusy(false);
    }
  }

  function canSyncToWishlist(productId: string) {
    return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(productId);
  }

  async function toggleWishlist(productId: string) {
    if (!account || !canSyncToWishlist(productId)) return;
    if (wishlistProductIDs.has(productId)) {
      await removeWishlistItem(productId);
      setWishlistItems((current) => current.filter((item) => item.product_id !== productId));
      return;
    }
    await addWishlistItem(productId);
    setWishlistItems((current) => [{ account_id: account.id, product_id: productId, added_at: new Date().toISOString() }, ...current]);
  }

  async function add(productId: string) {
    if (!account) {
      setGuestCart((current) => ({ ...current, [productId]: (current[productId] ?? 0) + 1 }));
      return;
    }
    const product = productByID.get(productId);
    if (!product) return;
    setCartBusy(true);
    try {
      const next = await addCartItem({ productID: productId, quantity: 1, unitPriceCents: toCents(product.price) });
      setCartData(next);
    } finally {
      setCartBusy(false);
    }
  }

  async function remove(productId: string) {
    if (!account) {
      setGuestCart((current) => {
        const next = { ...current };
        const quantity = (next[productId] ?? 0) - 1;
        if (quantity <= 0) delete next[productId];
        else next[productId] = quantity;
        return next;
      });
      return;
    }
    const quantity = (effectiveCart[productId] ?? 0) - 1;
    setCartBusy(true);
    try {
      const next = await updateCartItem(productId, Math.max(quantity, 0));
      setCartData(next);
    } finally {
      setCartBusy(false);
    }
  }

  async function submitCheckout() {
    if (!account) {
      setCheckoutNotice("Sign in first so the cart, order and delivery history can sync to your account.");
      return;
    }
    if (!cartItems.length) {
      setCheckoutNotice("Your cart is empty right now.");
      return;
    }
    setCheckoutBusy(true);
    setCheckoutNotice("");

    let orderID = "";
    let paymentSettled = false;
    let shipmentCreated = false;
    let cartFinalized = false;

    try {
      const order = await createOrder({
        cartID: cartData?.id,
        currency: "USD",
        shippingCents: toCents(deliveryPrice),
        deliveryMethod: deliveryOptions[delivery].title,
        deliveryAddress: checkoutForm.address,
        customerNote: checkoutForm.note,
        items: cartItems.map((product) => ({
          productID: product.id,
          quantity: effectiveCart[product.id],
          unitPriceCents: toCents(product.price)
        }))
      });
      orderID = order.id;

      const transactionRef = `demo-${Date.now()}`;
      const payment = await createPayment({
        orderID: order.id,
        provider: "demo-stripe",
        method: checkoutForm.paymentMethod,
        currency: order.currency,
        amountCents: order.total_cents,
        transactionRef
      });
      await confirmPayment(payment.id, payment.transaction_ref ?? transactionRef);
      paymentSettled = true;
      await updateOrderStatus(order.id, "paid");

      await createShipment({
        orderID: order.id,
        carrier: "Nova Poshta",
        serviceLevel: deliveryOptions[delivery].title,
        trackingNumber: `EM-${Date.now()}`,
        destinationAddress: checkoutForm.address,
        eta: new Date(Date.now() + 36 * 60 * 60 * 1000).toISOString()
      });
      shipmentCreated = true;
      await updateOrderStatus(order.id, "processing");

      await checkoutCart();
      cartFinalized = true;

      await loadOperationalData();
      setCheckoutNotice(`Order ${order.id.slice(0, 8)} is now checked out, paid and pushed into shipping.`);
    } catch (error) {
      if (orderID) {
        try {
          if (!paymentSettled) {
            await updateOrderStatus(orderID, "cancelled");
          }
        } catch {
          // Keep the original checkout error visible even if recovery is partial.
        }
      }
      await loadOperationalData().catch(() => undefined);

      const message = error instanceof Error ? error.message : "Checkout failed";
      if (orderID && paymentSettled && !shipmentCreated) {
        setCheckoutNotice(`${message}. Payment for order ${orderID.slice(0, 8)} succeeded, but shipment creation still needs attention.`);
      } else if (orderID && paymentSettled && shipmentCreated && !cartFinalized) {
        setCheckoutNotice(`${message}. Order ${orderID.slice(0, 8)} is already paid and queued for shipping, but cart finalization still needs attention.`);
      } else {
        setCheckoutNotice(message);
      }
    } finally {
      setCheckoutBusy(false);
    }
  }

  return (
    <main>
      <header className="site-shell nav">
        <a className="brand" href="#top" aria-label="Embedded Market home">
          <span className="brand-mark"><CircuitBoard size={21} /></span>
          Embedded Market
        </a>
        <nav className="nav-links" aria-label="Primary navigation">
          <a href="#catalog">Catalog</a>
          <a href="#ops">Account</a>
          <a href="#delivery">Delivery</a>
          <a href="#reviews">Orders</a>
        </nav>
        <button className="nav-cart account-chip" type="button" aria-label="Account status">
          <UserRound size={18} />
          <span>{account ? account.username : authLoading ? "Session" : "Sign in"}</span>
        </button>
      </header>

      <section className="hero" id="top">
        <img src="/images/embedded-hero.png" alt="Embedded electronics modules on a modern workbench" />
        <div className="hero-overlay" />
        <div className="site-shell hero-content">
          <div className="hero-copy">
            <span className="eyebrow"><Sparkles size={16} /> Gateway-linked storefront</span>
            <h1>Build firmware weekends that actually ship.</h1>
            <p>
              Live catalog, account login, synced cart, checkout orchestration and post-order shipment/payment visibility,
              all flowing through the gateway you already own.
            </p>
            <div className="hero-actions">
              <a className="button primary" href="#catalog">Shop modules <ChevronRight size={18} /></a>
              <a className="button ghost" href="#ops">Open account flow</a>
            </div>
          </div>
          <div className="hero-rail" aria-label="Store highlights">
            <div><strong>{account ? "Live" : "Guest"}</strong><span>account mode</span></div>
            <div><strong>{cartItems.length}</strong><span>items ready to check out</span></div>
            <div><strong>{orderBundle.orders.length}</strong><span>orders tracked</span></div>
          </div>
        </div>
      </section>

      <section className="site-shell stats-strip" id="analytics">
        {[ 
          { icon: Cpu, value: String(storefront.products.length), label: storefront.live ? "catalog modules live" : "demo modules ready" },
          { icon: Zap, value: String(storefront.recommendations.length), label: "gateway recommendation picks" },
          { icon: ShieldCheck, value: storefront.averageRating.toFixed(1), label: `${storefront.totalReviews} review signals` },
          { icon: Boxes, value: String(orderBundle.orders.length || storefront.products.filter((item) => item.oldPrice).length), label: account ? "orders tracked" : "discounted build kits" }
        ].map((item) => {
          const Icon = item.icon;
          return <div key={item.label}><Icon /><strong>{item.value}</strong><span>{item.label}</span></div>;
        })}
      </section>

      <section className="site-shell integration-note">
        <span className={storefront.live ? "status-pill live" : "status-pill"}>{storefront.live ? "Gateway live" : "Demo fallback"}</span>
        <p>{storefront.note}</p>
      </section>

      <section className="site-shell ops-grid" id="ops">
        <article className="ops-panel auth-panel">
          <div className="panel-head">
            <span className="eyebrow dark"><UserRound size={15} /> Account</span>
            {account ? <button className="text-action" type="button" onClick={() => void signOut()} disabled={authBusy}><LogOut size={16} /> Sign out</button> : null}
          </div>
          {account ? (
            <div className="account-summary">
              <strong>{account.username}</strong>
              <span>{account.email}</span>
              <small>{account.email_verified ? "Email verified" : "Email pending"} · {account.status}</small>
            </div>
          ) : (
            <>
              <div className="tab-row">
                <button className={authMode === "login" ? "chip active" : "chip"} type="button" onClick={() => setAuthMode("login")}>Login</button>
                <button className={authMode === "register" ? "chip active" : "chip"} type="button" onClick={() => setAuthMode("register")}>Register</button>
              </div>
              <div className="form-grid">
                {authMode === "register" ? <label className="field"><span>Email</span><input value={authForm.email} onChange={(event) => setAuthForm((current) => ({ ...current, email: event.target.value }))} /></label> : null}
                {authMode === "register" ? <label className="field"><span>Username</span><input value={authForm.username} onChange={(event) => setAuthForm((current) => ({ ...current, username: event.target.value }))} /></label> : null}
                {authMode === "login" ? <label className="field"><span>Email or username</span><input value={authForm.identifier} onChange={(event) => setAuthForm((current) => ({ ...current, identifier: event.target.value }))} /></label> : null}
                <label className="field"><span>Password</span><input type="password" value={authForm.password} onChange={(event) => setAuthForm((current) => ({ ...current, password: event.target.value }))} /></label>
              </div>
              {authError ? <div className="notice error">{authError}</div> : null}
              <button className="button primary wide" type="button" onClick={() => void submitAuth()} disabled={authBusy || authLoading}>
                {authBusy ? <LoaderCircle size={18} className="spin" /> : null}
                {authMode === "login" ? "Enter workspace" : "Create builder account"}
              </button>
            </>
          )}
        </article>

        <article className="ops-panel cart-live-panel">
          <div className="panel-head">
            <span className="eyebrow dark"><ShoppingBag size={15} /> Live cart</span>
            {cartBusy || opsLoading ? <span className="mini-status">Syncing...</span> : null}
          </div>
          <div className="cart-items compact">
            {cartItems.length ? cartItems.map((product) => (
              <div className="cart-item" key={product.id}>
                <span>{effectiveCart[product.id]}x</span>
                <div>
                  <strong>{product.name}</strong>
                  <small>{product.delivery}</small>
                </div>
                <em>${(product.price * effectiveCart[product.id]).toFixed(2)}</em>
              </div>
            )) : <div className="empty-state">Cart is waiting for the first module.</div>}
          </div>
          <div className="cart-total"><span>Subtotal</span><strong>${subtotal.toFixed(2)}</strong></div>
          <div className="cart-total"><span>Delivery</span><strong>{deliveryOptions[delivery].price}</strong></div>
          <div className="cart-total grand"><span>Total</span><strong>${total.toFixed(2)}</strong></div>
        </article>

        <article className="ops-panel checkout-panel">
          <div className="panel-head">
            <span className="eyebrow dark"><WalletCards size={15} /> Checkout flow</span>
          </div>
          <div className="form-grid">
            <label className="field"><span>Delivery address</span><textarea value={checkoutForm.address} onChange={(event) => setCheckoutForm((current) => ({ ...current, address: event.target.value }))} rows={3} /></label>
            <label className="field"><span>Customer note</span><textarea value={checkoutForm.note} onChange={(event) => setCheckoutForm((current) => ({ ...current, note: event.target.value }))} rows={2} /></label>
            <label className="field"><span>Payment method</span>
              <select value={checkoutForm.paymentMethod} onChange={(event) => setCheckoutForm((current) => ({ ...current, paymentMethod: event.target.value }))}>
                <option value="card">Card</option>
                <option value="apple-pay">Apple Pay</option>
                <option value="invoice">Invoice</option>
              </select>
            </label>
          </div>
          {checkoutNotice ? <div className="notice">{checkoutNotice}</div> : null}
          <button className="button checkout wide" type="button" onClick={() => void submitCheckout()} disabled={checkoutBusy || authLoading}>
            {checkoutBusy ? <LoaderCircle size={18} className="spin" /> : null}
            Create order, payment and shipment
          </button>
        </article>
      </section>

      <section className="site-shell catalog-layout" id="catalog">
        <div className="catalog-main">
          <div className="section-heading">
            <div>
              <span className="eyebrow dark"><SlidersHorizontal size={15} /> Curated catalog</span>
              <h2>Parts that make prototypes feel expensive.</h2>
            </div>
            <label className="search-box">
              <Search size={18} />
              <input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Search ESP32, OLED, sensors..."
              />
            </label>
          </div>

          <div className="category-row" aria-label="Product categories">
            {storefront.categories.map((item) => (
              <button className={item.value === category ? "chip active" : "chip"} key={item.value} onClick={() => setCategory(item.value)} type="button">
                {item.label}
              </button>
            ))}
          </div>

          {loading ? <div className="loading-banner">Loading storefront data from the gateway...</div> : null}

          <div className="product-grid">
            {visibleProducts.map((product, index) => (
              <article className={`product-card ${product.accent}`} key={product.id} style={{ "--delay": `${index * 70}ms` } as CSSProperties}>
                <button className={wishlistProductIDs.has(product.id) ? "favorite active" : "favorite"} type="button" aria-label={wishlistProductIDs.has(product.id) ? `Remove ${product.name} from wishlist` : `Save ${product.name}`} onClick={() => void toggleWishlist(product.id)} disabled={!account || !canSyncToWishlist(product.id)}>
                  <Heart size={17} />
                </button>
                <div className="product-photo">
                  <img src={product.imageUrl} alt={product.name} style={{ objectPosition: product.imagePosition }} />
                  <span>{product.tag}</span>
                </div>
                <div className="product-body">
                  <div className="rating"><Star size={15} fill="currentColor" /> {product.rating} <span>({product.reviews})</span></div>
                  <h3>{product.name}</h3>
                  <p>{product.specs.join(" / ")}</p>
                  <div className="stock-line"><BadgeCheck size={15} /> {product.stock}</div>
                  <div className="product-footer">
                    <div className="price">
                      <strong>${product.price}</strong>
                      {product.oldPrice ? <span>${product.oldPrice}</span> : null}
                    </div>
                    <div className="quantity">
                      {effectiveCart[product.id] ? <button type="button" onClick={() => void remove(product.id)} aria-label={`Remove ${product.name}`}><Minus size={15} /></button> : null}
                      {effectiveCart[product.id] ? <span>{effectiveCart[product.id]}</span> : null}
                      <button type="button" onClick={() => void add(product.id)} aria-label={`Add ${product.name}`}><Plus size={15} /></button>
                    </div>
                  </div>
                </div>
              </article>
            ))}
          </div>
        </div>

        <aside className="cart-panel" aria-label="Cart summary">
          <div className="cart-head">
            <span><ShoppingBag size={18} /> Checkout rail</span>
            <strong>${total.toFixed(2)}</strong>
          </div>
          <div className="cart-items">
            {cartItems.map((product) => (
              <div className="cart-item" key={product.id}>
                <span>{effectiveCart[product.id]}x</span>
                <div>
                  <strong>{product.name}</strong>
                  <small>{product.delivery}</small>
                </div>
                <em>${(product.price * effectiveCart[product.id]).toFixed(2)}</em>
              </div>
            ))}
          </div>
          <div className="cart-total"><span>Delivery</span><strong>{deliveryOptions[delivery].price}</strong></div>
          <div className="cart-total grand"><span>Total</span><strong>${total.toFixed(2)}</strong></div>
          <button className="button checkout" type="button" onClick={() => void submitCheckout()} disabled={checkoutBusy}>{account ? "Run live checkout" : "Sign in to checkout"}</button>
        </aside>
      </section>

      <section className="delivery-band" id="delivery">
        <div className="site-shell delivery-grid">
          <div>
            <span className="eyebrow dark"><Truck size={15} /> Delivery brain</span>
            <h2>Pick the route before the soldering iron is hot.</h2>
            <p>
              Delivery choices flow directly into order and shipment creation. Once checkout fires, the selected route becomes the shipping service level.
            </p>
          </div>
          <div className="delivery-options">
            {deliveryOptions.map((option, index) => {
              const Icon = option.icon;
              return (
                <button className={delivery === index ? "delivery-option active" : "delivery-option"} key={option.title} onClick={() => setDelivery(index)} type="button">
                  <Icon size={22} />
                  <span>
                    <strong>{option.title}</strong>
                    <small>{option.time}</small>
                  </span>
                  <em>{option.price}</em>
                </button>
              );
            })}
          </div>
        </div>
      </section>

      <section className="site-shell lower-grid" id="reviews">
        <div className="signal-panel">
          <span className="eyebrow dark"><Star size={15} /> Reviews</span>
          <h2>Builders keep receipts.</h2>
          <blockquote>
            “Now the storefront does not just look premium. It can authenticate, sync the cart, place an order,
            mint a payment and create a shipment trail through your own services.”
          </blockquote>
          <div className="reviewer">
            <span>DX</span>
            <div><strong>Developer experience</strong><small>{account ? "Account session active" : "Guest mode active"}</small></div>
          </div>
        </div>
        <div className="signal-panel recommendations">
          <span className="eyebrow dark"><Sparkles size={15} /> Recommendations</span>
          <h2>Trending through recommendation-service</h2>
          {storefront.recommendations.map((product) => (
            <div className="recommendation" key={product.id}>
              <span>{product.category}</span>
              <strong>{product.name}</strong>
              <em>${product.price}</em>
            </div>
          ))}
        </div>
        <div className="signal-panel timeline">
          <span className="eyebrow dark"><Clock3 size={15} /> Orders</span>
          <h2>{latestOrder ? `Latest order ${latestOrder.id.slice(0, 8)}` : "Order stream waiting"}</h2>
          {latestOrder ? (
            <ol>
              <li><strong>{latestOrder.status}</strong><span>{latestOrder.delivery_method}</span></li>
              <li><strong>{latestPayment ? latestPayment.status : "payment pending"}</strong><span>{latestPayment ? latestPayment.provider : "Payment service idle"}</span></li>
              <li><strong>{latestShipment ? latestShipment.status : "shipment pending"}</strong><span>{latestShipment ? latestShipment.carrier : "Shipping service idle"}</span></li>
            </ol>
          ) : (
            <div className="empty-state">Place the first real order and this panel becomes your fulfillment pulse.</div>
          )}
        </div>
      </section>
    </main>
  );

  async function hydrateProducts(
    items: CatalogProduct[],
    categories: CatalogCategory[],
    signal: AbortSignal
  ): Promise<Product[]> {
    const categoryByID = new Map(categories.map((item) => [item.id, item]));
    const summaries = await Promise.all(items.map((item) => getReviewSummary(item.id).catch(() => null)));
    if (signal.aborted) return [];

    return items.map((item, index) => {
      const summary = summaries[index];
      return mapCatalogProduct(item, categoryByID.get(item.category_id), summary, index);
    });
  }
}

function mapCatalogProduct(
  item: CatalogProduct,
  category: CatalogCategory | undefined,
  summary: ReviewSummary | null,
  index: number
): Product {
  const price = derivePrice(item, index);
  const specs = item.specs.slice(0, 3).map((spec) => spec.value || spec.key).filter(Boolean);
  return {
    id: item.id,
    name: item.name,
    category: category?.name ?? "Embedded parts",
    tag: item.featured ? "Featured" : item.sku || "Build-ready",
    price,
    oldPrice: item.featured ? price + 11 : undefined,
    rating: summary?.average_rating ? Number(summary.average_rating.toFixed(1)) : 4.6 + (index % 3) * 0.1,
    reviews: summary?.total_reviews ?? 18 + index * 7,
    stock: item.status === "active" ? `${60 + index * 13} in stock` : "Backorder batch",
    accent: accentCycle[index % accentCycle.length],
    imagePosition: imagePositions[index % imagePositions.length],
    delivery: deliveryCycle[index % deliveryCycle.length],
    specs: specs.length ? specs : [item.short_description || "Prototype ready", item.status || "Active"],
    imageUrl: pickImageURL(item)
  };
}

function derivePrice(item: CatalogProduct, index: number) {
  if (item.sku) {
    const digits = item.sku.replace(/\D/g, "");
    if (digits) {
      const parsed = Number(digits.slice(-3));
      if (parsed > 0) {
        return Math.max(16, Math.min(149, parsed));
      }
    }
  }
  return 24 + index * 11;
}

function pickImageURL(item: CatalogProduct) {
  if (item.image_url) return item.image_url;
  const mediaImage = item.media.find((entry) => entry.type === "image" || entry.url);
  return mediaImage?.url || "/images/embedded-hero.png";
}

function mapRecommendations(items: Recommendation[], products: Product[]) {
  if (!items.length) return products.slice(1, 4);
  const byID = new Map(products.map((product) => [product.id, product]));
  const mapped = items.map((item) => byID.get(item.product_id)).filter((item): item is Product => Boolean(item));
  return mapped.length ? mapped.slice(0, 3) : products.slice(1, 4);
}

function toCents(amount: number) {
  return Math.round(amount * 100);
}

export default App;
