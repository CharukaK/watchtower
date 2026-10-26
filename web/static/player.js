const video = document.querySelector('#video');
const message = document.querySelector('#message');
let reader;
let defaultControls = false;
const mediaURL = new URL(`/media/channel-${document.body.dataset.channelId}/`, window.location.href);

function queryBoolean(name, fallback) {
  const value = new URLSearchParams(window.location.search).get(name);
  if (value === null) return fallback;
  return ['1', 'yes', 'true'].includes(value.toLowerCase());
}

function setMessage(text) {
  message.textContent = text;
  video.controls = text ? false : defaultControls;
}

function startPlayer() {
  video.controls = queryBoolean('controls', true);
  video.muted = queryBoolean('muted', true);
  video.autoplay = queryBoolean('autoplay', true);
  video.playsInline = queryBoolean('playsinline', true);
  video.disablePictureInPicture = queryBoolean('disablepictureinpicture', false);
  defaultControls = video.controls;

  reader = new MediaMTXWebRTCReader({
    url: new URL(`whep${window.location.search}`, mediaURL).toString(),
    onError: setMessage,
    onTrack: (event) => {
      setMessage('');
      video.srcObject = event.streams[0];
    },
    onDataChannel: (event) => {
      event.channel.binaryType = 'arraybuffer';
    },
  });
}

const readerScript = document.createElement('script');
readerScript.src = new URL('reader.js', mediaURL).toString();
readerScript.addEventListener('load', startPlayer);
readerScript.addEventListener('error', () => setMessage('Unable to reach the media server'));
document.head.append(readerScript);

window.addEventListener('beforeunload', () => reader?.close());
