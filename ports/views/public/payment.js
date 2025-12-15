// Initialize Stripe.js
const stripe = Stripe("pk_test_51STUa8Bg6i8TN2jAtL8kc8s0UOrrWLSE1V5ENqILnX9Oejnc8PBlNRecFq4DseXblO23WIH0ibr3JVoaPkbJCxQY00YtLAYmm8");

initialize();

// Fetch Checkout Session and retrieve the client secret
async function initialize() {
  const fetchClientSecret = async () => {
    const response = await fetch("/api/checkout", {
      method: "POST",
    });
    const { clientSecret } = await response.json();
    return clientSecret;
  };

  // Initialize Checkout
  const checkout = await stripe.initEmbeddedCheckout({
    fetchClientSecret,
  });

  // Mount Checkout
  checkout.mount("#checkout");
}