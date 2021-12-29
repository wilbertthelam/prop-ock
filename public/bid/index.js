const numberOnlyReg = /^\d+$/;

(() => {
  const urlSearchParams = new URLSearchParams(window.location.search);
  const params = Object.fromEntries(urlSearchParams.entries());

  const playerId = params["playerId"] || "";
  const auctionId = params["auctionId"] || "";
  const senderPsId = params["senderPsId"] || "";

  hideById = (id) => {
    document.getElementById(id).style.display = "none";
  };

  showById = (id) => {
    document.getElementById(id).style.display = "block";
  };

  isInvalidBidNotInteger = (currentBid) =>
    currentBid !== "" && !numberOnlyReg.test(currentBid);

  isInvalidBidNegative = (currentBid) => currentBid && Number(currentBid) < 0;

  // If any of the bid validators fail, it is an invalid bid
  isInvalidBid = (currentBid) =>
    [isInvalidBidNotInteger, isInvalidBidNegative].some((fn) => fn(currentBid));

  onBidInputType = (event) => {
    const currentBid = event.target.value;

    if (isInvalidBidNotInteger(currentBid)) {
      showById("invalid-bid-not-integer");
    } else {
      hideById("invalid-bid-not-integer");
    }

    if (isInvalidBidNegative(currentBid)) {
      showById("invalid-bid-negative");
    } else {
      hideById("invalid-bid-negative");
    }

    if (isInvalidBid(currentBid) || currentBid === "") {
      document.getElementById("bid-button").setAttribute("disabled", "true");
    } else {
      document.getElementById("bid-button").removeAttribute("disabled");
    }
  };

  onBidInputKeypress = (event) => {
    if (event.keyCode === 13) {
      onPlaceBid();
    }
  };

  getPlayer = (playerId) => {
    const xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = () => {
      if (xhttp.readyState === XMLHttpRequest.DONE && xhttp.status == 200) {
        const { name, position, team, id, image } = JSON.parse(
          xhttp.responseText
        );

        document.getElementById("player-name").innerHTML = name;
        document.getElementById("player-position").innerHTML = position;
        document.getElementById("player-team").innerHTML = team;
        document.getElementById("player-image").setAttribute("src", image);
      }
    };
    xhttp.open("GET", `/api/player?playerId=${playerId}`, true);
    xhttp.send();
  };

  onPlaceBid = () => {
    const currentBid = document.getElementById("bid-input").value;
    if (Number(currentBid) >= 0) {
      console.log("sending bid: " + currentBid);

      const xhttp = new XMLHttpRequest();
      xhttp.open("POST", `/api/auction/bid`, true);
      xhttp.setRequestHeader("Content-type", "application/json");
      xhttp.onreadystatechange = () => {
        if (xhttp.readyState === XMLHttpRequest.DONE && xhttp.status == 200) {
          showById("bid-placed-success");
        }
      };
      xhttp.send(
        JSON.stringify({
          auction_id: auctionId,
          sender_ps_id: senderPsId,
          player_id: playerId,
          bid: Number(currentBid),
        })
      );
    }
  };

  getPlayer(playerId);

  window.onload = () => {
    // Handle when the user submits a bid
    document.getElementById("bid-button").addEventListener("click", onPlaceBid);

    document
      .getElementById("bid-input")
      .addEventListener("input", onBidInputType);

    document
      .getElementById("bid-input")
      .addEventListener("keypress", onBidInputKeypress);
  };
})();
