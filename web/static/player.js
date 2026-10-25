const video = document.querySelector('#video');
const message = document.querySelector('#message');
let reader;
let defaultControls = false;

function queryBoolean(name, fallback) {
  const value = new URLSearchParams(window.location.search).get(name);
  if (value === null) return fallback;
  return ['1', 'yes', 'true'].includes(value.toLowerCase());
}

function setMessage(text) {
  message.textContent = text;
  video.controls = text ? false : defaultControls;
}

window.addEventListener('load', () => {
  video.controls = queryBoolean('controls', true);
  video.muted = queryBoolean('muted', true);
  video.autoplay = queryBoolean('autoplay', true);
  video.playsInline = queryBoolean('playsinline', true);
  video.disablePictureInPicture = queryBoolean('disablepictureinpicture', false);
  defaultControls = video.controls;

  reader = new MediaMTXWebRTCReader({
    url: document.body.dataset.whepUrl + window.location.search,
    onError: setMessage,
    onTrack: (event) => {
      setMessage('');
      video.srcObject = event.streams[0];
    },
    onDataChannel: (event) => {
      event.channel.binaryType = 'arraybuffer';
    },
  });
});

window.addEventListener('beforeunload', () => reader?.close());
