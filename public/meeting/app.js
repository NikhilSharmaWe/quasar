window.addEventListener("DOMContentLoaded", (_) => { 
    path = window.location.pathname;
    key = path.replace("/meets/",'').replace("/",'');
    let mk = document.getElementById("meeting_key");
    mk.innerHTML = `<p><strong>Meeting ID: </strong>${key}</p>`;

    navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {

      const chatInput = document.getElementById('chat-input');
      const chatButton = document.getElementById('send-button');
      const chatMessages = document.getElementById('chat-messages');

      chatButton.addEventListener('click', () => {
        const message = chatInput.value;
        if (message.trim() !== '') {
          ws.send(JSON.stringify({ event: 'chat', data: message }));
          chatInput.value = '';
        }
      });

      const config = {
        codec: 'vp8',
        iceServers: [
            {
                "urls": "stun:stun.l.google.com:19302",
            },
        ]
      }
      let pc = new RTCPeerConnection(config)

      pc.ontrack = function (event) {
        if (event.track.kind === 'audio' || event.track.kind === 'video') {
            let streamID = event.streams[0].id;
            const pElement = document.createElement('p');
            pElement.style.color = 'white';
            pElement.id = event.streams[0].id; // Use the streamID as the id
            document.getElementById('remoteVideos').appendChild(pElement);

            let el = document.createElement(event.track.kind);
            el.srcObject = event.streams[0];
            el.autoplay = true;
            el.controls = true;
            pElement.appendChild(el);

            event.track.onmute = function(event) {
                el.play();
            }

            event.streams[0].onremovetrack = ({track}) => {
              if (el.parentNode) {
                el.parentNode.removeChild(el);
              }
              let e = document.getElementById(streamID);
              e.remove();
            }
        }
      }

      pc.onremovestream = function (event) {
        console.log("removing")
        const streamId = event.stream.id;
        const element = remoteVideoElements[streamId];
        if (element) {
          element.remove();
        }
      }


      document.getElementById('localVideo').srcObject = stream
      stream.getTracks().forEach(track => pc.addTrack(track, stream))
      console.log("local video data:", stream);
      let ws = new WebSocket("ws://" + window.location.host + "/websocket");
      pc.onicecandidate = e => {
        if (!e.candidate) {
          return
        }

        ws.send(JSON.stringify({event: 'candidate', data: JSON.stringify(e.candidate)}))
      }

      ws.onclose = function(evt) {
        console.log("websocket connection closed")
        window.alert("Websocket has closed")
      }

      ws.onmessage = function(evt) {
        let msg = JSON.parse(evt.data)
        if (!msg) {
          return console.log('failed to parse msg')
        }

        console.log(msg);
        // console.log(participantInfo);

        switch (msg.event) {
          case 'offer':
            let offer = JSON.parse(msg.data)
            if (!offer) {
              return console.log('failed to parse answer')
            }
            console.log("Recived offer")
            pc.setRemoteDescription(offer)
            pc.createAnswer().then(answer => {
              pc.setLocalDescription(answer)
              ws.send(JSON.stringify({event: 'answer', data: JSON.stringify(answer)}))
            })
            return

          case 'candidate':
            let candidate = JSON.parse(msg.data)
            if (!candidate) {
              return console.log('failed to parse candidate')
            }

            pc.addIceCandidate(candidate)
            return
          
          case 'chat':
            const chatData = msg.data;
            const message = chatData.Message;
            const username = chatData.Username;

            console.log(username + ":" + message)

            const messageElement = document.createElement('div');
            messageElement.innerHTML = `<p><strong>${username}:</strong> ${message}</p>`;
            chatMessages.appendChild(messageElement);
            return

          case 'my_username':
            let un = document.getElementById("your_username");
            un.innerHTML = msg.data;
            return

          case 'participant':
            const participantData = msg.data;
            // participantInfo[participantData.StreamID] = participantData.Username;
            let p = document.getElementById(participantData.StreamID);
            p.innerHTML = participantData.Username;
            // console.log(participantInfo);
            return
        }
      }

      ws.onerror = function(evt) {
        console.log("ERROR: " + evt.data)
      }
    }).catch(window.alert)
});
