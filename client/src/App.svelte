<script lang="ts">
  import { Styles } from "sveltestrap";
  import {
    Button,
    Card,
    CardBody,
    CardTitle,
    CardText,
    Col,
    Icon,
    Image,
    InputGroup,
    InputGroupText,
    Input,
    Row,
  } from "sveltestrap";
  import { onMount } from "svelte";

  let bidAmountInputVal = "";

  const player = {
    id: "",
    name: "",
    position: "",
    team: "",
    image: "",
  };

  let bidAmount = -1;

  const urlSearchParams = new URLSearchParams(window.location.search);
  const params = Object.fromEntries(urlSearchParams.entries());

  const {
    player_id: playerId,
    auction_id: auctionId,
    sender_ps_id: senderPsId,
  } = params;

  const getBid = async (
    auctionId: string,
    playerId: string,
    senderPsId: string
  ) => {
    const getBidResponse = await fetch(
      `/api/auction/bid?player_id=${playerId}&auction_id=${auctionId}&sender_ps_id=${senderPsId}`
    );
    if (!getBidResponse.ok) {
      alert("failed to get bid");
      return;
    }

    bidAmount = await getBidResponse.json();
  };

  const getPlayer = async (playerId: string) => {
    const getPlayerResponse = await fetch(`/api/player?player_id=${playerId}`);
    if (!getPlayerResponse.ok) {
      alert("failed to get player");
      return;
    }

    const { id, name, position, team, image } = await getPlayerResponse.json();
    player.id = id;
    player.name = name;
    player.position = position;
    player.team = team;
    player.image = image;
  };

  onMount(async () => {
    getPlayer(playerId);
    getBid(auctionId, playerId, senderPsId);
  });

  const onPlaceBid = async () => {
    if (Number(bidAmountInputVal) >= 0) {
      const reqBody = {
        auction_id: auctionId,
        sender_ps_id: senderPsId,
        player_id: playerId,
        bid: Number(bidAmountInputVal),
      };
      const placeBidResponse = await fetch(`/api/auction/bid/make`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(reqBody),
      });
      if (!placeBidResponse.ok) {
        alert("failed to place bid");
        return;
      }

      bidAmount = Number(bidAmountInputVal);
    }
  };

  const onCancelBid = async () => {
    const reqBody = {
      auction_id: auctionId,
      sender_ps_id: senderPsId,
      player_id: playerId,
    };
    const placeBidResponse = await fetch(`/api/auction/bid/cancel`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(reqBody),
    });
    if (!placeBidResponse.ok) {
      alert("failed to cancel place bid");
      return;
    }

    // Reset bid amount to -1 once canceled
    bidAmount = -1;
  };

  const onBidInputKeypress = (event) => {
    if (event.keyCode === 13) {
      onPlaceBid();
    }
  };

  const onBidInput = (event) => {
    bidAmountInputVal = event.target.value;
  };

  export let name: string;
</script>

<Styles />

<main>
  <Row>
    <Col sm="12" md={{ size: 4, offset: 4 }}>
      <h1>Hello {name}!</h1>
      <Card color="light">
        <CardBody>
          <Image class="card-img-top" alt="player image" src={player.image} />
          <CardTitle>
            <h4>{player.name}</h4>
          </CardTitle>
          <CardText>
            <b>{player.position}</b> | {player.team}
          </CardText>
          {#if bidAmount < 0}
            <InputGroup size="lg">
              <InputGroupText>$</InputGroupText>
              <Input
                on:keypress={onBidInputKeypress}
                on:input={onBidInput}
                placeholder="Bid amount"
                min={0}
                type="number"
                step="1"
              />
              <InputGroupText>.00</InputGroupText>
              <Button color="primary" on:click={onPlaceBid}>Bid</Button>
            </InputGroup>
          {:else}
            <div>You currently have a bid out for ${bidAmount}</div>
            <Button color="danger" on:click={onCancelBid}
              ><Icon name="x-circle" /><span class="m-2">Cancel</span></Button
            >
          {/if}
        </CardBody>
      </Card>
    </Col>
  </Row>
</main>

<style>
  main {
    text-align: center;
    padding: 1em;
    max-width: 240px;
    margin: 0 auto;
  }

  h1 {
    color: #e6bd09;
    text-transform: uppercase;
    font-size: 2em;
    font-weight: 200;
  }

  @media (min-width: 640px) {
    main {
      max-width: none;
    }
  }
</style>
