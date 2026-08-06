const API_BASE = "http://localhost:8080/api/v1";

async function login() {

    const email = document.getElementById("email").value;

    const password = document.getElementById("password").value;

    if (!email || !password) {

        alert("Enter email and password");

        return;
    }

    const response = await fetch(`${API_BASE}/auth/login`, {

        method: "POST",

        headers: {

            "Content-Type": "application/json"
        },

        body: JSON.stringify({

            email,

            password
        })
    });

    const result = await response.json();

    console.log(result);

    if (!response.ok) {

        alert(result.error);

        return;
    }

    localStorage.setItem(

        "token",

        result.data.token
    );

    document.getElementById("status").innerHTML =
        "✅ Logged in successfully";

    alert("Login Successful");
}

async function payNow() {

    const orderId = document
        .getElementById("orderId")
        .value;

    if (!orderId) {

        alert("Enter Order ID");

        return;
    }

    const token = localStorage.getItem("token");

    if (!token) {

        alert("Please login first.");

        return;
    }

    const response = await fetch(`${API_BASE}/payments`, {

        method: "POST",

        headers: {

            "Content-Type": "application/json",

            "Authorization": `Bearer ${token}`
        },

        body: JSON.stringify({

            order_id: orderId
        })
    });

    const result = await response.json();

    console.log(result);

    if (!response.ok) {

        alert(result.error);

        return;
    }

    const payment = result.data;

    const options = {

        key: payment.key_id,

        order_id: payment.gateway_order_id,

        amount: payment.amount * 100,

        currency: payment.currency,

        name: "Enterprise Order Processing",

        description: "Order Payment",

        image: "https://razorpay.com/favicon.png",

        prefill: {

            email,

            contact: "9999999999"
        },

        notes: {

            order_id: payment.order_id
        },

        theme: {

            color: "#2563eb"
        },

        modal: {

            ondismiss: function () {

                console.log("Checkout Closed");
            }
        },

        handler: function (response) {

            console.log(response);

            alert("Payment Successful!");
        }
    };

    const razorpay = new Razorpay(options);

    razorpay.open();
}