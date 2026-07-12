export type CatalogCategory = {
  id: string;
  name: string;
  slug: string;
  description: string;
};

export type CatalogProduct = {
  id: string;
  category_id: string;
  brand_id: string;
  slug: string;
  sku: string;
  name: string;
  short_description: string;
  description: string;
  datasheet_url: string;
  image_url: string;
  status: string;
  featured: boolean;
  specs: Array<{ key: string; value: string }>;
  media: Array<{ url: string; type: string; sort_order: number }>;
};

export type ReviewSummary = {
  product_id: string;
  total_reviews: number;
  average_rating: number;
};

export type Recommendation = {
  product_id: string;
  score: number;
  reason: string;
};

export type Account = {
  id: string;
  email: string;
  username: string;
  status: string;
  email_verified: boolean;
  created_at: string;
};

export type AuthResponse = {
  account: Account;
  token_type: string;
  access_token: string;
  access_token_expires_at: string;
  refresh_token: string;
  refresh_token_expires_at: string;
};

export type CartItem = {
  id: string;
  product_id: string;
  quantity: number;
  unit_price_cents: number;
  line_total_cents: number;
};

export type Cart = {
  id: string;
  account_id: string;
  status: string;
  currency: string;
  subtotal_cents: number;
  items: CartItem[];
};

export type OrderItem = {
  id: string;
  product_id: string;
  quantity: number;
  unit_price_cents: number;
  line_total_cents: number;
};

export type Order = {
  id: string;
  account_id: string;
  cart_id?: string;
  status: string;
  currency: string;
  subtotal_cents: number;
  shipping_cents: number;
  total_cents: number;
  delivery_method: string;
  delivery_address: string;
  customer_note?: string;
  items: OrderItem[];
  created_at: string;
};

export type Payment = {
  id: string;
  order_id: string;
  account_id: string;
  status: string;
  provider: string;
  method: string;
  currency: string;
  amount_cents: number;
  transaction_ref?: string;
  paid_at?: string;
  updated_at: string;
};

export type Shipment = {
  id: string;
  order_id: string;
  account_id: string;
  status: string;
  carrier: string;
  service_level: string;
  tracking_number?: string;
  destination_address: string;
  eta?: string;
  updated_at: string;
};

export type WishlistItem = {
  account_id: string;
  product_id: string;
  added_at: string;
};

type Paginated<T> = {
  items: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
};

const API_BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? "";
const ACCESS_TOKEN_KEY = "embedded-market-access-token";

let accessTokenCache = readStoredAccessToken();
let refreshInFlight: Promise<AuthResponse> | null = null;

function readStoredAccessToken() {
  if (typeof window === "undefined") return "";
  return window.localStorage.getItem(ACCESS_TOKEN_KEY) ?? "";
}

export function getAccessToken() {
  return accessTokenCache;
}

export function setAccessToken(token: string | null) {
  accessTokenCache = token ?? "";
  if (typeof window === "undefined") return;
  if (token) {
    window.localStorage.setItem(ACCESS_TOKEN_KEY, token);
  } else {
    window.localStorage.removeItem(ACCESS_TOKEN_KEY);
  }
}

async function performRefresh() {
  if (!refreshInFlight) {
    refreshInFlight = requestJSON<AuthResponse>("/api/v1/auth/refresh", {
      method: "POST",
      body: JSON.stringify({}),
      skipAuthRetry: true
    }).finally(() => {
      refreshInFlight = null;
    });
  }

  const result = await refreshInFlight;
  setAccessToken(result.access_token);
  return result;
}

async function requestJSON<T>(
  path: string,
  init?: RequestInit & { authenticated?: boolean; skipAuthRetry?: boolean }
): Promise<T> {
  const headers = new Headers(init?.headers ?? {});
  if (!headers.has("Content-Type") && init?.body) {
    headers.set("Content-Type", "application/json");
  }
  if (init?.authenticated && accessTokenCache) {
    headers.set("Authorization", `Bearer ${accessTokenCache}`);
  }

  const response = await fetch(`${API_BASE}${path}`, {
    credentials: "include",
    ...init,
    headers
  });

  if (response.status === 401 && init?.authenticated && !init.skipAuthRetry) {
    try {
      await performRefresh();
    } catch {
      setAccessToken(null);
      throw new Error("Your session expired. Please sign in again.");
    }
    return requestJSON<T>(path, { ...init, skipAuthRetry: true });
  }

  if (!response.ok) {
    let message = `Request failed with status ${response.status}`;
    try {
      const payload = await response.json();
      message = payload?.error?.message ?? message;
    } catch {
      // Keep the generic message when the body is not JSON.
    }
    throw new Error(message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export function listCatalogProducts(params: { query?: string; categorySlug?: string }) {
  const search = new URLSearchParams({ page_size: "12", featured: "true" });
  if (params.query) search.set("q", params.query);
  if (params.categorySlug) search.set("category_slug", params.categorySlug);
  return requestJSON<Paginated<CatalogProduct>>(`/api/v1/catalog/products?${search.toString()}`);
}

export function listCategories() {
  return requestJSON<{ items: CatalogCategory[] }>("/api/v1/catalog/categories");
}

export function listTrendingRecommendations(limit = 4) {
  return requestJSON<{ items: Recommendation[] }>(`/api/v1/recommendations/trending?limit=${limit}`);
}

export function getReviewSummary(productID: string) {
  return requestJSON<ReviewSummary>(`/api/v1/reviews/summary/${productID}`);
}

export function registerAccount(input: { email: string; username: string; password: string }) {
  return requestJSON<AuthResponse>("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function loginAccount(input: { identifier: string; password: string }) {
  return requestJSON<AuthResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function refreshSession() {
  return performRefresh();
}

export function logoutSession() {
  return requestJSON<void>("/api/v1/auth/logout", {
    method: "POST",
    body: JSON.stringify({})
  });
}

export function getCurrentAccount() {
  return requestJSON<Account>("/api/v1/auth/me", { authenticated: true });
}

export function getActiveCart() {
  return requestJSON<Cart>("/api/v1/cart", { authenticated: true });
}

export function listWishlist() {
  return requestJSON<Paginated<WishlistItem>>("/api/v1/wishlist?page_size=50", { authenticated: true });
}

export function addWishlistItem(productID: string) {
  return requestJSON<{ status: string }>(`/api/v1/wishlist/${productID}`, {
    method: "POST",
    authenticated: true
  });
}

export function removeWishlistItem(productID: string) {
  return requestJSON<void>(`/api/v1/wishlist/${productID}`, {
    method: "DELETE",
    authenticated: true
  });
}

export function addCartItem(input: { productID: string; quantity: number; unitPriceCents: number }) {
  return requestJSON<Cart>("/api/v1/cart/items", {
    method: "POST",
    authenticated: true,
    body: JSON.stringify({
      product_id: input.productID,
      quantity: input.quantity,
      unit_price_cents: input.unitPriceCents
    })
  });
}

export function updateCartItem(productID: string, quantity: number) {
  return requestJSON<Cart>(`/api/v1/cart/items/${productID}`, {
    method: "PUT",
    authenticated: true,
    body: JSON.stringify({ quantity })
  });
}

export function clearCart() {
  return requestJSON<Cart>("/api/v1/cart/items", {
    method: "DELETE",
    authenticated: true
  });
}

export function checkoutCart() {
  return requestJSON<Cart>("/api/v1/cart/checkout", {
    method: "POST",
    authenticated: true
  });
}

export function listOrders() {
  return requestJSON<Paginated<Order>>("/api/v1/orders?page_size=10", { authenticated: true });
}

export function createOrder(input: {
  cartID?: string;
  currency: string;
  shippingCents: number;
  deliveryMethod: string;
  deliveryAddress: string;
  customerNote: string;
  items: Array<{ productID: string; quantity: number; unitPriceCents: number }>;
}) {
  return requestJSON<Order>("/api/v1/orders", {
    method: "POST",
    authenticated: true,
    body: JSON.stringify({
      cart_id: input.cartID,
      currency: input.currency,
      shipping_cents: input.shippingCents,
      delivery_method: input.deliveryMethod,
      delivery_address: input.deliveryAddress,
      customer_note: input.customerNote,
      items: input.items.map((item) => ({
        product_id: item.productID,
        quantity: item.quantity,
        unit_price_cents: item.unitPriceCents
      }))
    })
  });
}

export function updateOrderStatus(orderID: string, status: string) {
  return requestJSON<Order>(`/api/v1/orders/${orderID}/status`, {
    method: "PUT",
    authenticated: true,
    body: JSON.stringify({ status })
  });
}

export function listPayments() {
  return requestJSON<Paginated<Payment>>("/api/v1/payments?page_size=10", { authenticated: true });
}

export function createPayment(input: { orderID: string; provider: string; method: string; currency: string; amountCents: number; transactionRef: string }) {
  return requestJSON<Payment>("/api/v1/payments", {
    method: "POST",
    authenticated: true,
    body: JSON.stringify({
      order_id: input.orderID,
      provider: input.provider,
      method: input.method,
      currency: input.currency,
      amount_cents: input.amountCents,
      transaction_ref: input.transactionRef
    })
  });
}

export function confirmPayment(paymentID: string, transactionRef: string) {
  return requestJSON<Payment>(`/api/v1/payments/${paymentID}/confirm`, {
    method: "POST",
    authenticated: true,
    body: JSON.stringify({ transaction_ref: transactionRef })
  });
}

export function listShipments() {
  return requestJSON<Paginated<Shipment>>("/api/v1/shipping/shipments?page_size=10", { authenticated: true });
}

export function createShipment(input: { orderID: string; carrier: string; serviceLevel: string; trackingNumber: string; destinationAddress: string; eta?: string }) {
  return requestJSON<Shipment>("/api/v1/shipping/shipments", {
    method: "POST",
    authenticated: true,
    body: JSON.stringify({
      order_id: input.orderID,
      carrier: input.carrier,
      service_level: input.serviceLevel,
      tracking_number: input.trackingNumber,
      destination_address: input.destinationAddress,
      eta: input.eta
    })
  });
}

function ensureSessionID() {
  const key = "embedded-market-session";
  const existing = window.localStorage.getItem(key);
  if (existing) return existing;
  const created = typeof crypto !== "undefined" && "randomUUID" in crypto ? crypto.randomUUID() : `session-${Date.now()}`;
  window.localStorage.setItem(key, created);
  return created;
}

export function trackPageView(path: string) {
  return requestJSON<void>("/api/v1/analytics/events", {
    method: "POST",
    body: JSON.stringify({
      session_id: ensureSessionID(),
      event_type: "page_view",
      path,
      referrer: document.referrer,
      query: window.location.search,
      user_agent: navigator.userAgent
    })
  });
}
