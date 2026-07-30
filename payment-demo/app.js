const API_BASE = "http://localhost:8080/api/v1";

async function payNow() {

    const orderId = document
        .getElementById("orderId")
        .value;

    if (!orderId) {

        alert("Enter Order ID");

        return;
    }

    const response = await fetch(`${API_BASE}/payments`, {

        method: "POST",

        headers: {

            "Content-Type": "application/json"
        },

        body: JSON.stringify({

            order_id: orderId
        })
    });

    const result = await response.json();

    console.log(result);

    if (!result.success) {

        alert(result.error);

        return;
    }

    const payment = result.data;

    const options = {

        key: payment.key_id,

        amount: payment.amount * 100,

        currency: payment.currency,

        order_id: payment.gateway_order_id,

        name: "Enterprise Order Processing",

        description: "Order Payment",

        handler: function (response) {

            console.log(response);

            alert("Payment Successful!");

        },

        prefill: {

            email: "test@example.com",

            contact: "9999999999"
        },

        theme: {

            color: "#2563eb"
        }
    };

    const razorpay = new Razorpay(options);

    razorpay.open();
}