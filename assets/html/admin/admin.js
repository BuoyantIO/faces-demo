'use strict';

const POLL_INTERVAL   = 3000;
const WIDGET_INTERVAL = 2000;

const EMOJI_CATEGORIES = [
  { name: 'Faces', icon: '😀', emojis: [
    '😀','😃','😄','😁','😆','😅','🤣','😂','🙂','🙃','😉','😊','😇','🥰','😍','🤩','😘','😗','😚','😙','🥲',
    '😋','😛','😜','🤪','😝','🤑','🤗','🤭','🤫','🤔','🤐','🤨','😐','😑','😶','😶‍🌫️','😏','😒','🙄','😬','🤥',
    '😌','😔','😪','🤤','😴','😷','🤒','🤕','🤢','🤮','🤧','🥵','🥶','🥴','😵','😵‍💫','🤯','🤠','🥳','🥸',
    '😎','🤓','🧐','😕','😟','🙁','☹️','😮','😯','😲','😳','🥺','🥹','😦','😧','😨','😰','😥','😢','😭',
    '😱','😖','😣','😞','😓','😩','😫','🥱','😤','😡','😠','🤬','😈','👿','💀','☠️','💩','🤡','👹','👺',
    '👻','👽','👾','🤖','😺','😸','😹','😻','😼','😽','🙀','😿','😾',
  ]},
  { name: 'People', icon: '👋', emojis: [
    '👋','🤚','🖐️','✋','🖖','🫱','🫲','🫳','🫴','🤙','💪','🦾','🖕','✌️','🤞','🫰','🤟','🤘','🤙','👈','👉',
    '👆','👇','☝️','👍','👎','✊','👊','🤛','🤜','👏','🙌','🫶','👐','🤲','🤝','🙏','✍️','💅','🤳','🦵','🦶',
    '👂','🦻','👃','🫦','🫀','🫁','🧠','🦷','🦴','👀','👁️','👅','👶','🧒','👦','👧','🧑','👱','👨','🧔',
    '👩','🧓','👴','👵','👮','👷','💂','🕵️','👩‍⚕️','👨‍⚕️','👩‍🌾','👨‍🌾','👩‍🍳','👨‍🍳','👩‍🎓','👨‍🎓',
    '👩‍🎤','👨‍🎤','👩‍🏫','👨‍🏫','👩‍🏭','👨‍🏭','👩‍💻','👨‍💻','👩‍💼','👨‍💼','👩‍🔧','👨‍🔧',
    '👩‍🚒','👨‍🚒','👩‍✈️','👨‍✈️','👩‍🚀','👨‍🚀','👩‍⚖️','👨‍⚖️','🧙','🧚','🧛','🧜','🧝','🧞','🧟','🧌',
  ]},
  { name: 'Animals', icon: '🐶', emojis: [
    '🐶','🐱','🐭','🐹','🐰','🦊','🐻','🐼','🐨','🐯','🦁','🐮','🐷','🐽','🐸','🐵','🙈','🙉','🙊',
    '🐔','🐧','🐦','🐤','🦆','🦅','🦉','🦇','🐺','🐗','🐴','🦄','🐝','🪱','🐛','🦋','🐌','🐞','🐜',
    '🪲','🦟','🦗','🕷️','🕸️','🦂','🐢','🐍','🦎','🦕','🦖','🐙','🦑','🦐','🦞','🦀','🐡','🐠','🐟',
    '🐬','🐳','🐋','🦈','🐊','🐅','🐆','🦓','🦍','🦧','🦣','🐘','🦛','🦏','🐪','🐫','🦒','🦘','🦬',
    '🐃','🐂','🐄','🐎','🐖','🐏','🐑','🦙','🐐','🦌','🐕','🐩','🦮','🐕‍🦺','🐈','🐈‍⬛','🪶','🐓',
    '🦃','🦤','🦚','🦜','🦢','🦩','🕊️','🐇','🦝','🦨','🦡','🦫','🦦','🦥','🐁','🐀','🐿️','🦔',
    '🌵','🎄','🌲','🌳','🌴','🌱','🌿','☘️','🍀','🎋','🎍','🍃','🍂','🍁','🍄','🌾','💐','🌷','🌹','🌺','🌸','🌼','🌻',
  ]},
  { name: 'Food', icon: '🍕', emojis: [
    '🍏','🍎','🍐','🍊','🍋','🍌','🍉','🍇','🍓','🫐','🍈','🍒','🍑','🥭','🍍','🥥','🥝','🍅',
    '🍆','🥑','🥦','🥬','🥒','🌶️','🫑','🥕','🧄','🧅','🥔','🍠','🫚','🫛','🧀','🥚','🍳','🧈',
    '🥞','🧇','🥓','🥩','🍗','🍖','🦴','🌭','🍔','🍟','🍕','🫓','🥙','🧆','🌮','🌯','🫔','🥗',
    '🥘','🫕','🥫','🍝','🍜','🍲','🍛','🍣','🍱','🥟','🦪','🍤','🍙','🍚','🍘','🍥','🥮','🍢',
    '🧁','🍰','🎂','🍮','🍭','🍬','🍫','🍿','🍩','🍪','🌰','🥜','🍯','🧃','🥤','🧋','☕','🫖',
    '🍵','🍶','🍺','🍻','🥂','🍷','🫗','🥃','🍸','🍹','🧉','🍾','🧊','🥄','🍴','🍽️','🥢','🫙',
  ]},
  { name: 'Travel', icon: '🚀', emojis: [
    '🚗','🚕','🚙','🚌','🚎','🏎️','🚓','🚑','🚒','🚐','🛻','🚚','🚛','🚜','🛵','🏍️','🛺','🚲',
    '🛴','🛹','🛼','🚏','⛽','🚨','🚥','🚦','🛑','🚧','⚓','🛟','⛵','🚤','🛥️','🛳️','⛴️','🚢',
    '✈️','🛩️','🛫','🛬','🪂','🚁','🛸','🚀','🛶','🚠','🚡','🚂','🚃','🚄','🚅','🚆','🚇','🚈',
    '🚉','🚊','🚝','🚞','🚋','🚍','🚑','🚒','🏔️','⛰️','🌋','🗻','🏕️','🏖️','🏜️','🏝️','🏞️',
    '🏟️','🏛️','🏗️','🏠','🏡','🏢','🏣','🏤','🏥','🏦','🏨','🏩','🏪','🏫','🏭','🗼','🗽',
    '⛪','🕌','🛕','⛩️','🕍','💒','🏰','🏯','🗺️','🌐','🗾','🧭','🌍','🌎','🌏','🪐','☀️','🌤️',
    '⛅','🌥️','☁️','🌦️','🌧️','⛈️','🌩️','🌨️','❄️','☃️','⛄','🌬️','🌀','🌈','⚡','🔥','💧','🌊','🌁','🌫️',
  ]},
  { name: 'Activities', icon: '⚽', emojis: [
    '⚽','🏀','🏈','⚾','🥎','🎾','🏐','🏉','🥏','🎱','🪀','🏓','🏸','🏒','🏑','🥍','🏏','🪃',
    '🥅','⛳','🪁','🎣','🤿','🥊','🥋','🎽','🛹','🛼','🛷','🥌','🏋️','🤸','🤼','🤺','🏇',
    '⛷️','🏂','🪂','🤼','🤸','🤺','🏊','🚵','🚴','🤾','🏌️','🧘','🧗','🤺','🤼','⛹️','🤽',
    '🎯','🎳','🎰','🧩','🪆','🎭','🎪','🤹','🎨','🖼️','🎬','🎤','🎧','🎼','🎹','🪘','🥁',
    '🎷','🎺','🎸','🪕','🎻','🪗','🎮','🕹️','🎲','♟️','🎯','🎳','🎰','🧸','🪅','🎊','🎉','🎈','🎁','🎀','🎗️','🎫','🎟️','🏆','🥇','🥈','🥉','🏅','🎖️',
  ]},
  { name: 'Objects', icon: '💡', emojis: [
    '💡','🔦','🕯️','🪔','🔋','🔌','💻','🖥️','🖨️','⌨️','🖱️','🖲️','💾','💿','📀','🧮',
    '📱','☎️','📞','📟','📠','📺','📻','🧭','⏱️','⏲️','⏰','🕰️','⌛','⏳','📡','🔭','🔬',
    '🩺','💉','🩹','💊','🩻','🩼','🦽','🦾','🪜','🧲','🔧','🔩','🪛','🪚','🔨','⛏️','⚒️',
    '🛠️','🗜️','🔑','🗝️','🔐','🔏','🔒','🔓','🔫','🪃','🏹','⚔️','🛡️','🪤','🧰','🪣',
    '🪝','🧲','🚪','🪞','🪟','🛋️','🪑','🚿','🛁','🪠','🧴','🧷','🧹','🧺','🧻','🧼',
    '🪥','🛒','📦','📫','📪','📬','📭','📮','🗳️','✏️','✒️','🖊️','🖋️','📝','📓','📔',
    '📒','📕','📗','📘','📙','📚','📖','🔖','🏷️','💰','💴','💵','💶','💷','💸','💳','🧾','💎','⚖️','🔮','🪄','🧸',
  ]},
  { name: 'Symbols', icon: '❤️', emojis: [
    '❤️','🧡','💛','💚','💙','💜','🖤','🤍','🤎','💔','❤️‍🔥','❤️‍🩹','💕','💞','💓','💗','💖','💘','💝','💟',
    '☮️','✝️','☪️','🕉️','☸️','✡️','🔯','🕎','☯️','☦️','🛐','⛎','♈','♉','♊','♋','♌','♍','♎','♏','♐','♑','♒','♓',
    '⚜️','🔰','✅','☑️','✔️','❎','❌','❓','❔','❕','❗','‼️','⁉️','🚫','⛔','📵','🔞','💯',
    '🔅','🔆','📶','🔇','🔈','🔉','🔊','📢','📣','🔔','🔕','💬','💭','🗯️','💤','🔱','⚜️','♻️',
    '✨','⭐','🌟','💫','⚡','🔥','💥','❄️','💧','🌊','🌈','☀️','🌙','☁️','⛅','🌤️',
    '➕','➖','➗','✖️','♾️','🔁','🔂','▶️','⏩','⏭️','⏯️','◀️','⏪','⏮️','⬆️','⬇️','⬅️','➡️',
    '↗️','↘️','↙️','↖️','↕️','↔️','↩️','↪️','⤴️','⤵️','🔄','🔃','🔀','🔼','🔽',
    '🆕','🆓','🆒','🆗','🆙','🆘','🆚','🆎','🆑','🅰️','🅱️','🔤','🔡','🔢','🔣',
  ]},
];

const COLOR_PALETTE = [
  // Original palette
  { name: 'blue',     hex: '#66CCEE' },
  { name: 'green',    hex: '#228833' },
  { name: 'darkblue', hex: '#4477AA' },
  { name: 'yellow',   hex: '#CCBB44' },
  { name: 'red',      hex: '#EE6677' },
  { name: 'purple',   hex: '#AA3377' },
  { name: 'grey',     hex: '#BBBBBB' },
  // Popular additions
  { name: 'orange',   hex: '#FF8800' },
  { name: 'pink',     hex: '#FF66AA' },
  { name: 'teal',     hex: '#00AAAA' },
  { name: 'lime',     hex: '#88CC00' },
  { name: 'coral',    hex: '#FF5544' },
  { name: 'indigo',   hex: '#5555CC' },
  { name: 'white',    hex: '#FFFFFF' },
];

// Searchable keyword map for all emojis in EMOJI_CATEGORIES.
// Keywords are space-separated; search matches any substring.
const EMOJI_NAMES = {
  // Faces
  '😀':'grinning smile happy','😃':'grinning smile happy open','😄':'laugh grin smile happy',
  '😁':'beam grin smile happy','😆':'laugh squint smile happy','😅':'sweat smile nervous',
  '🤣':'rofl rolling laugh funny hilarious','😂':'joy tear laugh funny cry',
  '🙂':'smile slight','🙃':'upside smile flip','😉':'wink','😊':'blush smile',
  '😇':'angel halo innocent smiling','🥰':'hearts love smiling','😍':'heart eyes love',
  '🤩':'star struck excited wow','😘':'kiss blow wink heart','😗':'kiss','😚':'kiss eyes closed',
  '😙':'kiss smile','🥲':'tear smile grateful','😋':'yum lick food tasty',
  '😛':'tongue out playful','😜':'wink tongue playful','🤪':'zany crazy wacky',
  '😝':'squint tongue','🤑':'money eyes dollar rich','🤗':'hug arms open',
  '🤭':'mouth secret oops','🤫':'shush quiet secret shhh','🤔':'thinking ponder hmm',
  '🤐':'zipper mouth silent quiet','🤨':'eyebrow raised skeptical suspicious',
  '😐':'neutral expressionless','😑':'expressionless blank','😶':'no mouth silent',
  '😶‍🌫️':'fog cloud silent','😏':'smirk sly','😒':'unamused annoyed','🙄':'eye roll',
  '😬':'grimace nervous awkward','🤥':'lie pinocchio','😌':'relieved calm',
  '😔':'pensive sad','😪':'sleepy tired','🤤':'drool','😴':'sleeping tired zzz sleep',
  '😷':'mask sick ill face','🤒':'sick fever thermometer','🤕':'hurt bandage injury',
  '🤢':'nausea sick green','🤮':'vomit sick','🤧':'sneeze sick sniff',
  '🥵':'hot sweat fire sweating','🥶':'cold freeze blue freezing',
  '🥴':'woozy dizzy drunk','😵':'dizzy spiral faint','😵‍💫':'dizzy spiral',
  '🤯':'explode mind blown shocked','🤠':'cowboy hat western','🥳':'party celebrate hat',
  '🥸':'disguise glasses mustache','😎':'cool sunglasses','🤓':'nerd glasses smart',
  '🧐':'monocle curious fancy','😕':'confused puzzled','😟':'worried concerned',
  '🙁':'sad frown slight','☹️':'sad frown unhappy','😮':'surprise open mouth',
  '😯':'hushed surprised','😲':'astonished shocked gasp','😳':'flushed embarrassed blush',
  '🥺':'pleading puppy eyes','🥹':'tear holding smile grateful',
  '😦':'frown open mouth','😧':'anguish','😨':'fear scared anxious',
  '😰':'anxious sweat nervous','😥':'sad relieved','😢':'cry sad tear weep',
  '😭':'sob cry loud tear sad weep','😱':'scream horror fear',
  '😖':'confound frustrate','😣':'persevere struggle','😞':'disappoint sad',
  '😓':'sweat downcast tired','😩':'weary tired','😫':'tired weary exhausted',
  '🥱':'yawn tired sleepy','😤':'steam triumph angry','😡':'pout angry red mad',
  '😠':'angry mad','🤬':'swear curse angry','😈':'devil evil smile purple',
  '👿':'angry devil evil','💀':'skull dead death','☠️':'skull crossbones death',
  '💩':'poop poo','🤡':'clown','👹':'ogre monster','👺':'goblin',
  '👻':'ghost boo','👽':'alien extraterrestrial','👾':'alien monster pixel game',
  '🤖':'robot','😺':'cat grinning smile','😸':'cat grin big eyes',
  '😹':'cat joy tear laugh','😻':'cat heart eyes love','😼':'cat smirk wry',
  '😽':'cat kiss','🙀':'cat weary shocked','😿':'cat cry sad','😾':'cat pout angry',
  // People
  '👋':'wave hello goodbye hand','🤚':'raised back hand stop',
  '🖐️':'hand spread fingers','✋':'raised hand high five stop','🖖':'vulcan salute spock',
  '🫱':'handshake','🫲':'handshake','🫳':'palm down','🫴':'palm up',
  '🤙':'hang loose call','💪':'muscle flex strong bicep arm','🦾':'mechanical arm',
  '🖕':'middle finger rude','✌️':'peace victory two fingers','🤞':'crossed fingers luck',
  '🫰':'snap fingers','🤟':'love sign','🤘':'rock horns','👈':'point left',
  '👉':'point right','👆':'point up','👇':'point down','☝️':'index point up',
  '👍':'thumbs up good yes approve','👎':'thumbs down bad no disapprove',
  '✊':'fist raised','👊':'fist punch','🤛':'fist left','🤜':'fist right',
  '👏':'clap applause hands','🙌':'raised hands celebrate','🫶':'heart hands love',
  '👐':'open hands','🤲':'palms together','🤝':'handshake','🙏':'pray please thank folded',
  '✍️':'writing hand pen','💅':'nail polish','🤳':'selfie','🦵':'leg kick',
  '🦶':'foot','👂':'ear listen','🦻':'ear hearing aid','👃':'nose smell',
  '🫦':'lips','🫀':'heart organ','🫁':'lungs','🧠':'brain','🦷':'tooth',
  '🦴':'bone','👀':'eyes look see','👁️':'eye','👅':'tongue',
  '👶':'baby','🧒':'child','👦':'boy child','👧':'girl child',
  '🧑':'person','👱':'blonde person','👨':'man','🧔':'beard person','👩':'woman',
  '🧓':'older person','👴':'old man elderly','👵':'old woman elderly',
  '👮':'police officer','👷':'construction worker','💂':'guard','🕵️':'detective spy',
  // Animals
  '🐶':'dog puppy','🐱':'cat kitty','🐭':'mouse','🐹':'hamster','🐰':'rabbit bunny',
  '🦊':'fox','🐻':'bear','🐼':'panda','🐨':'koala','🐯':'tiger face',
  '🦁':'lion','🐮':'cow face','🐷':'pig face','🐽':'pig nose','🐸':'frog',
  '🐵':'monkey face','🙈':'see no evil monkey','🙉':'hear no evil monkey',
  '🙊':'speak no evil monkey','🐔':'chicken hen','🐧':'penguin',
  '🐦':'bird','🐤':'chick','🦆':'duck','🦅':'eagle bird','🦉':'owl',
  '🦇':'bat','🐺':'wolf','🐗':'boar pig','🐴':'horse face','🦄':'unicorn magic',
  '🐝':'bee honeybee','🪱':'worm','🐛':'bug caterpillar','🦋':'butterfly',
  '🐌':'snail slow','🐞':'ladybug beetle','🐜':'ant','🪲':'beetle',
  '🦟':'mosquito','🦗':'cricket','🕷️':'spider','🕸️':'cobweb spider',
  '🦂':'scorpion','🐢':'turtle tortoise slow','🐍':'snake','🦎':'lizard',
  '🦕':'dinosaur sauropod','🦖':'dinosaur t-rex','🐙':'octopus','🦑':'squid',
  '🦐':'shrimp','🦞':'lobster','🦀':'crab','🐡':'blowfish','🐠':'tropical fish',
  '🐟':'fish','🐬':'dolphin','🐳':'whale spout','🐋':'whale','🦈':'shark',
  '🐊':'crocodile','🐅':'tiger','🐆':'leopard','🦓':'zebra','🦍':'gorilla ape',
  '🦧':'orangutan ape','🦣':'mammoth','🐘':'elephant','🦛':'hippo rhinoceros',
  '🦏':'rhinoceros','🐪':'camel one hump','🐫':'camel two humps','🦒':'giraffe',
  '🦘':'kangaroo','🦬':'bison buffalo','🐃':'water buffalo','🐂':'ox bull',
  '🐄':'cow','🐎':'horse','🐖':'pig','🐏':'ram sheep','🐑':'sheep ewe',
  '🦙':'llama','🐐':'goat','🦌':'deer','🐕':'dog','🐩':'poodle dog','🦮':'guide dog',
  '🐕‍🦺':'service dog','🐈':'cat','🐈‍⬛':'black cat','🪶':'feather','🐓':'rooster chicken',
  '🦃':'turkey','🦤':'dodo bird','🦚':'peacock','🦜':'parrot bird','🦢':'swan bird',
  '🦩':'flamingo','🕊️':'dove peace bird','🐇':'rabbit bunny','🦝':'raccoon',
  '🦨':'skunk','🦡':'badger','🦫':'beaver','🦦':'otter','🦥':'sloth',
  '🐁':'mouse','🐀':'rat','🐿️':'chipmunk','🦔':'hedgehog',
  '🌵':'cactus desert','🎄':'christmas tree holiday','🌲':'tree evergreen pine',
  '🌳':'tree deciduous','🌴':'palm tree tropical','🌱':'seedling sprout new growth',
  '🌿':'herb plant leaf','☘️':'clover irish luck','🍀':'four leaf clover luck',
  '🎋':'bamboo','🎍':'pine','🍃':'leaf wind','🍂':'fallen leaf autumn',
  '🍁':'maple leaf autumn canada','🍄':'mushroom','🌾':'wheat grain sheaf',
  '💐':'bouquet flowers','🌷':'tulip flower','🌹':'rose flower love',
  '🌺':'hibiscus flower tropical','🌸':'blossom cherry flower spring',
  '🌼':'blossom flower yellow','🌻':'sunflower',
  // Food
  '🍏':'apple green fruit','🍎':'apple red fruit','🍐':'pear fruit','🍊':'orange fruit',
  '🍋':'lemon fruit yellow sour','🍌':'banana fruit yellow','🍉':'watermelon fruit',
  '🍇':'grapes fruit purple','🍓':'strawberry fruit red','🫐':'blueberry fruit',
  '🍈':'melon fruit','🍒':'cherry fruit red','🍑':'peach fruit','🥭':'mango fruit',
  '🍍':'pineapple fruit tropical','🥥':'coconut fruit','🥝':'kiwi fruit','🍅':'tomato',
  '🍆':'eggplant aubergine vegetable','🥑':'avocado','🥦':'broccoli vegetable',
  '🥬':'leafy green vegetable','🥒':'cucumber vegetable','🌶️':'pepper hot chili spicy',
  '🫑':'bell pepper vegetable','🥕':'carrot vegetable orange','🧄':'garlic',
  '🧅':'onion','🥔':'potato','🍠':'sweet potato','🫚':'olive oil',
  '🫛':'pea pod','🧀':'cheese','🥚':'egg','🍳':'egg pan frying cooking breakfast',
  '🧈':'butter','🥞':'pancakes breakfast','🧇':'waffle','🥓':'bacon',
  '🥩':'meat steak','🍗':'chicken drumstick','🍖':'meat bone',
  '🦴':'bone','🌭':'hot dog sausage','🍔':'burger hamburger','🍟':'fries french potato',
  '🍕':'pizza','🫓':'flatbread','🥙':'stuffed pita','🧆':'falafel',
  '🌮':'taco mexican','🌯':'burrito wrap','🫔':'tamale','🥗':'salad',
  '🥘':'paella shallow pan','🫕':'fondue pot','🥫':'canned food',
  '🍝':'spaghetti pasta','🍜':'noodles ramen soup','🍲':'pot stew',
  '🍛':'curry rice','🍣':'sushi fish rice','🍱':'bento box','🥟':'dumpling',
  '🦪':'oyster shellfish','🍤':'fried shrimp','🍙':'rice ball','🍚':'rice',
  '🍘':'rice cracker','🍥':'fish cake swirl','🥮':'mooncake','🍢':'oden',
  '🧁':'cupcake','🍰':'cake slice birthday','🎂':'birthday cake',
  '🍮':'custard pudding','🍭':'lollipop candy sweet','🍬':'candy sweet',
  '🍫':'chocolate','🍿':'popcorn','🍩':'donut doughnut','🍪':'cookie',
  '🌰':'chestnut','🥜':'peanut nut','🍯':'honey pot','🧃':'juice box drink',
  '🥤':'cup straw drink','🧋':'bubble tea boba','☕':'coffee hot drink warm',
  '🫖':'teapot tea','🍵':'tea hot drink matcha','🍶':'sake',
  '🍺':'beer mug ale','🍻':'beers toast cheers clinking','🥂':'champagne wine toast',
  '🍷':'wine red glass','🫗':'pouring liquid','🥃':'whiskey tumbler glass',
  '🍸':'cocktail martini','🍹':'tropical drink cocktail','🧉':'mate drink',
  '🍾':'bottle cork champagne','🧊':'ice cube cold','🥄':'spoon',
  '🍴':'fork knife cutlery','🍽️':'plate cutlery','🥢':'chopsticks','🫙':'jar',
  // Travel
  '🚗':'car automobile drive red','🚕':'taxi cab yellow','🚙':'car SUV drive',
  '🚌':'bus transport public','🚎':'trolleybus','🏎️':'racing car fast speed',
  '🚓':'police car','🚑':'ambulance hospital emergency','🚒':'fire truck engine',
  '🚐':'minibus van','🛻':'pickup truck','🚚':'truck delivery','🚛':'lorry trailer',
  '🚜':'tractor farm','🛵':'scooter moped','🏍️':'motorcycle bike','🛺':'auto rickshaw',
  '🚲':'bicycle bike pedal','🛴':'kick scooter','🛹':'skateboard','🛼':'roller skate',
  '🚏':'bus stop','⛽':'fuel gas pump','🚨':'siren light police',
  '🚥':'traffic light horizontal','🚦':'traffic light vertical','🛑':'stop sign',
  '🚧':'construction barrier','⚓':'anchor ship maritime','🛟':'life preserver ring',
  '⛵':'sailboat','🚤':'speedboat','🛥️':'motorboat','🛳️':'passenger ship',
  '⛴️':'ferry boat','🚢':'ship cruise','✈️':'airplane plane flight travel',
  '🛩️':'small airplane','🛫':'airplane takeoff','🛬':'airplane landing',
  '🪂':'parachute skydive','🚁':'helicopter','🛸':'UFO flying saucer alien',
  '🚀':'rocket space launch','🛶':'canoe rowboat','🚠':'mountain tramway',
  '🚡':'monorail','🚂':'steam locomotive train','🚃':'railway car',
  '🚄':'bullet train fast','🚅':'shinkansen bullet train','🚆':'train',
  '🚇':'metro subway underground','🚈':'light rail','🚉':'station',
  '🚊':'tram','🚝':'monorail','🚞':'mountain railway',
  '🏔️':'mountain snow peak alps','⛰️':'mountain','🌋':'volcano eruption',
  '🗻':'mount fuji japan','🏕️':'camping tent outdoors','🏖️':'beach sun sand',
  '🏜️':'desert sand dry','🏝️':'island tropical','🏞️':'national park',
  '🏟️':'stadium arena','🏛️':'classical building column','🏗️':'construction building',
  '🏠':'house home','🏡':'house garden home','🏢':'office building city',
  '🏣':'post office japan','🏤':'post office','🏥':'hospital medical',
  '🏦':'bank','🏨':'hotel','🏩':'love hotel','🏪':'store shop convenience',
  '🏫':'school','🏭':'factory industrial','🗼':'tokyo tower','🗽':'statue liberty usa',
  '⛪':'church','🕌':'mosque islam','🛕':'hindu temple','⛩️':'shinto shrine japan',
  '🕍':'synagogue','💒':'wedding chapel','🏰':'castle medieval',
  '🏯':'castle japanese','🗺️':'map world','🌐':'globe internet web',
  '🗾':'japan island','🧭':'compass navigation','🌍':'earth africa europe world',
  '🌎':'earth americas world','🌏':'earth asia world','🪐':'planet saturn space rings',
  '☀️':'sun sunny warm yellow','🌤️':'sun cloud partly sunny','⛅':'partly cloudy cloud sun',
  '🌥️':'cloud sun behind','☁️':'cloud overcast grey','🌦️':'rain sun partly',
  '🌧️':'rain cloud wet','⛈️':'thunderstorm rain lightning','🌩️':'lightning storm',
  '🌨️':'snow cloud flurry','❄️':'snowflake cold winter ice','☃️':'snowman',
  '⛄':'snowman winter cold','🌬️':'wind blow air','🌀':'cyclone tornado hurricane',
  '🌈':'rainbow colorful','⚡':'lightning bolt zap electric fast',
  '🔥':'fire flame hot burn','💧':'water drop rain','🌊':'wave ocean water sea',
  '🌁':'fog mist','🌫️':'fog haze',
  // Activities
  '⚽':'soccer football ball sport','🏀':'basketball ball sport','🏈':'american football ball',
  '⚾':'baseball ball','🥎':'softball ball','🎾':'tennis ball racket',
  '🏐':'volleyball ball','🏉':'rugby ball','🥏':'flying disc frisbee',
  '🎱':'pool billiards eight ball','🪀':'yo-yo','🏓':'table tennis ping pong',
  '🏸':'badminton shuttlecock','🏒':'hockey ice stick','🏑':'field hockey',
  '🥍':'lacrosse','🏏':'cricket bat ball','🪃':'boomerang',
  '🥅':'goal net','⛳':'golf hole flag','🪁':'slingshot','🎣':'fishing rod',
  '🤿':'diving mask snorkel','🥊':'boxing gloves fight punch','🥋':'martial arts karate judo',
  '🎽':'running shirt','🛹':'skateboard','🛼':'roller skate','🛷':'sled winter',
  '🥌':'curling stone','🏋️':'weightlifting gym','🤸':'cartwheel gymnastics',
  '🤼':'wrestling','🤺':'fencing sword','🏇':'horse racing',
  '⛷️':'skiing snow','🏂':'snowboarding','🪂':'parachute','🤼':'wrestling',
  '🏊':'swimming','🚵':'mountain biking','🚴':'cycling bike','🤾':'handball',
  '🏌️':'golf','🧘':'yoga meditation calm','🧗':'climbing',
  '⛹️':'basketball player','🤽':'water polo',
  '🎯':'target bullseye dart accuracy','🎳':'bowling','🎰':'slot machine casino',
  '🧩':'puzzle jigsaw piece','🪆':'matryoshka doll','🎭':'theater masks arts',
  '🎪':'circus tent','🤹':'juggling circus','🎨':'art palette paint brush',
  '🖼️':'frame picture art','🎬':'movie camera film clapper','🎤':'microphone sing karaoke',
  '🎧':'headphones music listen audio','🎼':'music score notes sheet',
  '🎹':'piano keyboard music','🪘':'drum bongo','🥁':'drums percussion',
  '🎷':'saxophone jazz music','🎺':'trumpet music','🎸':'guitar music rock',
  '🪕':'banjo country music','🎻':'violin strings music','🪗':'accordion',
  '🎮':'video game controller game play','🕹️':'joystick arcade game',
  '🎲':'dice game random','♟️':'chess strategy','🎯':'target',
  '🎳':'bowling','🎰':'slot','🧸':'teddy bear toy plush','🪅':'pinata',
  '🎊':'confetti celebrate party','🎉':'party popper celebrate tada',
  '🎈':'balloon party birthday','🎁':'gift present wrap','🎀':'ribbon bow gift',
  '🎗️':'ribbon awareness','🎫':'ticket event','🎟️':'admission ticket',
  '🏆':'trophy winner champion first','🥇':'gold medal first winner',
  '🥈':'silver medal second','🥉':'bronze medal third','🏅':'medal sport','🎖️':'military medal',
  // Objects
  '💡':'light bulb idea bright','🔦':'flashlight torch','🕯️':'candle flame',
  '🪔':'diya lamp','🔋':'battery power energy','🔌':'plug electric power',
  '💻':'laptop computer code','🖥️':'desktop computer monitor screen',
  '🖨️':'printer','⌨️':'keyboard type','🖱️':'mouse computer click',
  '🖲️':'trackball','💾':'floppy disk save','💿':'CD disc','📀':'DVD disc',
  '🧮':'abacus calculator','📱':'phone smartphone mobile cell',
  '☎️':'telephone phone call classic','📞':'phone handset call',
  '📟':'pager beeper','📠':'fax machine','📺':'television TV screen',
  '📻':'radio','🧭':'compass navigation direction','⏱️':'stopwatch timer',
  '⏲️':'timer clock','⏰':'alarm clock wake','🕰️':'mantel clock',
  '⌛':'hourglass time done','⏳':'hourglass sand time wait',
  '📡':'satellite antenna signal','🔭':'telescope space stars',
  '🔬':'microscope science lab biology','🩺':'stethoscope doctor',
  '💉':'syringe injection needle vaccine','🩹':'bandage plaster',
  '💊':'pill medicine drug tablet','🩻':'x-ray scan','🩼':'crutch',
  '🦽':'manual wheelchair','🦾':'mechanical arm','🪜':'ladder',
  '🧲':'magnet metal attract','🔧':'wrench tool fix repair',
  '🔩':'bolt nut screw','🪛':'screwdriver tool','🪚':'saw tool',
  '🔨':'hammer tool build','⛏️':'pickaxe mine','⚒️':'hammer pick',
  '🛠️':'tools hammer wrench fix','🗜️':'clamp compress','🔑':'key lock open',
  '🗝️':'old key','🔐':'locked key secure','🔏':'locked pen',
  '🔒':'locked padlock secure','🔓':'unlocked padlock open',
  '🔫':'water pistol squirt gun toy','🪃':'boomerang','🏹':'bow arrow',
  '⚔️':'crossed swords fight battle','🛡️':'shield protect defend',
  '🪤':'mousetrap','🧰':'toolbox repair','🪣':'bucket pail',
  '🪝':'hook hang','🧲':'magnet','🚪':'door entrance exit',
  '🪞':'mirror reflect','🪟':'window','🛋️':'couch sofa furniture',
  '🪑':'chair seat furniture','🚿':'shower bath clean',
  '🛁':'bathtub bath soak','🪠':'plunger unclog','🧴':'lotion bottle',
  '🧷':'safety pin','🧹':'broom sweep clean','🧺':'basket laundry',
  '🧻':'toilet roll paper','🧼':'soap wash clean','🪥':'toothbrush',
  '🛒':'shopping cart store','📦':'package box shipping',
  '📫':'mailbox closed mail','📪':'mailbox open','📬':'mailbox flag mail',
  '📭':'mailbox open flag','📮':'postbox red','🗳️':'ballot box vote',
  '✏️':'pencil write draw','✒️':'nib pen write','🖊️':'pen write','🖋️':'fountain pen',
  '📝':'memo note write','📓':'notebook school','📔':'notebook decorated',
  '📒':'ledger book','📕':'book red closed','📗':'book green',
  '📘':'book blue','📙':'book orange','📚':'books stack library',
  '📖':'book open read','🔖':'bookmark save','🏷️':'label tag price',
  '💰':'money bag cash','💴':'yen money','💵':'dollar bill money cash',
  '💶':'euro money','💷':'pound money','💸':'money flying wings',
  '💳':'credit card payment','🧾':'receipt bill','💎':'gem diamond jewel',
  '⚖️':'scales balance justice','🔮':'crystal ball magic fortune',
  '🪄':'magic wand','🧸':'teddy bear toy soft',
  // Symbols
  '❤️':'heart red love','🧡':'heart orange love','💛':'heart yellow love',
  '💚':'heart green love','💙':'heart blue love','💜':'heart purple love',
  '🖤':'heart black love dark','🤍':'heart white love','🤎':'heart brown love',
  '💔':'broken heart sad pain','❤️‍🔥':'heart fire love passion burning',
  '❤️‍🩹':'mending heart healing','💕':'two hearts love','💞':'revolving hearts love',
  '💓':'beating heart pulse love','💗':'growing heart love pink',
  '💖':'sparkling heart glitter love','💘':'heart arrow cupid love',
  '💝':'heart ribbon gift love','💟':'heart decoration',
  '☮️':'peace sign symbol','✝️':'cross christian','☪️':'crescent moon star islam',
  '🕉️':'om hinduism','☸️':'dharma wheel buddhism','✡️':'star david jewish',
  '🔯':'star dotted','🕎':'menorah hanukkah','☯️':'yin yang balance',
  '☦️':'orthodox cross','🛐':'worship prayer place','⛎':'ophiuchus zodiac',
  '♈':'aries','♉':'taurus','♊':'gemini','♋':'cancer','♌':'leo','♍':'virgo',
  '♎':'libra','♏':'scorpio','♐':'sagittarius','♑':'capricorn',
  '♒':'aquarius','♓':'pisces',
  '⚜️':'fleur-de-lis france gold','🔰':'japanese beginner green',
  '✅':'check done complete yes success','☑️':'checkbox check tick',
  '✔️':'checkmark done tick','❎':'cross mark no box','❌':'X cross error no delete',
  '❓':'question mark unknown','❔':'white question','❕':'exclamation white',
  '❗':'exclamation mark warning','‼️':'double exclamation','⁉️':'exclamation question',
  '🚫':'no prohibited forbidden banned','⛔':'no entry stop',
  '📵':'no mobile phone','🔞':'no under 18 adult','💯':'hundred percent perfect score',
  '🔅':'dim brightness low','🔆':'bright brightness high','📶':'signal bars wifi',
  '🔇':'muted sound off','🔈':'low sound volume','🔉':'medium sound volume',
  '🔊':'high sound loud','📢':'megaphone announce','📣':'loudspeaker',
  '🔔':'bell notification alert','🔕':'bell muted no notification',
  '💬':'speech bubble chat message','💭':'thought bubble thinking',
  '🗯️':'anger speech bubble','💤':'zzz sleep','🔱':'trident symbol',
  '♻️':'recycle green environment','✨':'sparkles glitter magic star shine',
  '⭐':'star yellow favorite','🌟':'glowing star shine','💫':'dizzy star spin',
  '⚡':'lightning bolt zap electric fast energy','🔥':'fire flame hot burn',
  '💥':'collision explosion boom bang','❄️':'snowflake cold winter ice',
  '💧':'water drop rain','🌊':'wave ocean water sea','🌈':'rainbow colorful',
  '☀️':'sun sunny warm','🌙':'moon night crescent dark','☁️':'cloud',
  '➕':'plus add math','➖':'minus subtract math','➗':'divide math',
  '✖️':'multiply times math','♾️':'infinity endless loop',
  '🔁':'repeat loop cycle refresh','🔂':'repeat once','▶️':'play start video',
  '⏩':'fast forward skip','⏭️':'next track skip','⏯️':'play pause',
  '◀️':'back reverse rewind','⏪':'rewind fast backward','⏮️':'previous track',
  '⬆️':'up arrow','⬇️':'down arrow','⬅️':'left arrow','➡️':'right arrow',
  '↗️':'up right arrow','↘️':'down right arrow','↙️':'down left arrow',
  '↖️':'up left arrow','↕️':'up down arrow','↔️':'left right arrow',
  '↩️':'return back arrow curved','↪️':'right curved arrow',
  '⤴️':'up right curved','⤵️':'down right curved',
  '🔄':'refresh repeat sync cycle','🔃':'clockwise arrows','🔀':'shuffle random',
  '🔼':'up button','🔽':'down button',
  '🆕':'new badge label','🆓':'free badge','🆒':'cool badge',
  '🆗':'ok badge','🆙':'up badge','🆘':'SOS emergency help danger',
  '🆚':'vs versus','🆎':'ab blood type','🆑':'cl button',
  '🅰️':'A blood type','🅱️':'B blood type',
  '🔤':'ABC letters input','🔡':'lowercase letters','🔢':'numbers input','🔣':'symbols',
};

// Returns emojis from all categories whose EMOJI_NAMES entry matches the query.
function searchEmojis(query) {
  const q = query.toLowerCase().trim();
  if (!q) return [];
  const results = [];
  const seen = new Set();
  for (const cat of window._allEmojiCategories || EMOJI_CATEGORIES) {
    if (!cat.emojis) continue;
    for (const emoji of cat.emojis) {
      if (seen.has(emoji)) continue;
      const name = EMOJI_NAMES[emoji] || '';
      if (name.includes(q) || emoji === q) {
        results.push(emoji);
        seen.add(emoji);
      }
    }
  }
  return results;
}

function renderEmojiSearchResults(emojis) {
  const grid = document.getElementById('emoji-picker-grid');
  grid.innerHTML = '';
  if (!emojis.length) {
    grid.innerHTML = '<div class="emoji-search-empty">No results — try a different word</div>';
    return;
  }
  for (const emoji of emojis) {
    const btn = document.createElement('button');
    btn.className = 'emoji-pick-btn';
    btn.textContent = emoji;
    btn.title = EMOJI_NAMES[emoji] || emoji;
    if (emoji === state.selectedSmiley) btn.classList.add('selected');
    btn.addEventListener('click', () => {
      clearEmojiGridSelection();
      btn.classList.add('selected');
      document.getElementById('custom-emoji-input').value = '';
      state.selectedSmiley = emoji;
      updateControlsPreview();
    });
    grid.appendChild(btn);
  }
}

// Tooltip text for pipeline nodes and metric cards
const TT = {
  publisher:   "Generates face messages (smiley + color) in background loops. Each message is written to MySQL first as a write-ahead guarantee, then pushed to the queue. Rate and pause state are controlled via Flow Controls.",
  mysql:       "Count of messages written to MySQL AND successfully pushed to the queue backend — the pipeline buffer from the database's perspective. Should closely match Queue Depth under normal conditions.",
  rabbitmq:    "Messages currently in RabbitMQ ready to be consumed by the subscriber. Should match DB Queued closely. A large gap means the max-depth cap was hit and the oldest queued messages were evicted.",
  redis:       "Messages currently in the Redis list ready to be consumed by the subscriber. Should match DB Queued closely. A large gap means LTRIM dropped the oldest messages.",
  subscriber:  "Pops messages from the queue and serves them to the faces-gui browser grid. Each successful pop acknowledges the row in MySQL. Pausing stops consumption — the queue fills up.",
  gui:         "The browser grid consuming faces. Each cell polls every 2 seconds. Delivered stops growing if no browser tabs are open, even while the queue is filling.",
  db_queued:   "MySQL rows with state='queued' — written to the database and pushed to the queue backend, awaiting the subscriber. This is the main pipeline buffer metric.",
  delivered:   "Total messages the subscriber has returned to the GUI (state='acknowledged' in MySQL). Grows only when a browser tab is actively polling and the subscriber is running.",
  depth:       "Current messages in the queue backend ready to be popped. Should track DB Queued closely. A significant gap between the two means the depth cap was exceeded and messages were evicted.",
  dropped:     "DB Queued minus Queue Depth. Non-zero means messages were evicted from the queue by the depth cap. They remain in MySQL as 'queued' and will be re-pushed to the queue on publisher restart.",
  pub_rate:    "Messages being published per second, as reported by the RabbitMQ management API. Reflects actual write throughput into the queue backend.",
  del_rate:    "Messages being delivered to the GUI per second, as reported by the RabbitMQ management API. If consistently lower than Pub Rate the queue will fill up over time.",
};

const RATE_TICKS = [
  { ms: 0    },
  { ms: 10   },
  { ms: 50   },
  { ms: 200  },
  { ms: 1000 },
];

const RATE_MAX = 1000;


// ── State ──────────────────────────────────────────────────────────────────
const state = {
  mode:           'classic',
  queueBackend:   'rabbitmq',
  lastPipeline:   null,
  lastControls:   null,
  lastConfig:     null,   // cached config for the poll indicator tooltip
  history:        [],
  selectedSmiley: null,
  selectedColor:  null,
  selectedColorHex: null,
  applyTarget:    'all',
  // Fault Injection global panel state
  lastInfra:       null,   // cached /api/infrastructure payload for scope targeting
  lastChaos:       null,   // cached /api/chaos payload for topology fault badges
  _fiTopoSig:      null,   // signature to skip redundant topology repaints
  fiScope:         'all',  // 'all' | 'zone' | 'node' | 'external' | 'onprem'
  fiScopeDetail:   null,   // selected zone name or node name when scope is zone/node
  fiSelectedServices: null, // array of service names; null = init to all-for-mode on first render
};

// ── Theme ──────────────────────────────────────────────────────────────────
function initTheme() {
  applyTheme(localStorage.getItem('faces-admin-theme') || 'dark');
  document.getElementById('theme-toggle').addEventListener('click', () => {
    const next = document.documentElement.dataset.theme === 'light' ? 'dark' : 'light';
    applyTheme(next);
    localStorage.setItem('faces-admin-theme', next);
  });
}
function applyTheme(theme) {
  document.documentElement.dataset.theme = theme;
  const icon  = document.getElementById('theme-icon');
  const label = document.getElementById('theme-label');
  if (icon)  icon.textContent  = theme === 'light' ? '🤓' : '😎';
  if (label) label.textContent = theme === 'light' ? '"Dork" mode' : 'Light mode';
}

// ── Sign Out ───────────────────────────────────────────────────────────────
async function signOut() {
  try { await fetch('/api/logout', { method: 'POST' }); } catch (_) {}
  window.location.href = '/login';
}

// ── Sidebar ────────────────────────────────────────────────────────────────
function initSidebar() {
  const sidebar  = document.getElementById('sidebar');
  const collapsed = localStorage.getItem('faces-admin-sidebar') === 'collapsed';
  if (collapsed) sidebar.classList.add('collapsed');

  function toggleSidebar() {
    sidebar.classList.toggle('collapsed');
    localStorage.setItem('faces-admin-sidebar',
      sidebar.classList.contains('collapsed') ? 'collapsed' : 'expanded');
  }

  document.getElementById('sidebar-toggle').addEventListener('click', toggleSidebar);
  document.getElementById('sidebar-brand').addEventListener('click', toggleSidebar);

  const VALID_PAGES = new Set(['overview', 'flow', 'liveview', 'faultinjection', 'controls', 'settings']);

  function navigateToPage(page) {
    if (!VALID_PAGES.has(page)) page = 'overview';
    document.querySelectorAll('.nav-item').forEach(i => i.classList.remove('active'));
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    const item = document.querySelector(`.nav-item[data-page="${page}"]`);
    const pageEl = document.getElementById('page-' + page);
    if (item)   item.classList.add('active');
    if (pageEl) pageEl.classList.add('active');
    document.querySelector('main').classList.toggle('lv-active', page === 'liveview');
    renderHeaderHealth();
    if (page === 'controls') refreshPodSelector();
    if (page === 'settings') { checkDBStatus(); checkQueueStatus(); pollEmojivotoSettings(); }
  }

  document.querySelectorAll('.nav-item[data-page]').forEach(item => {
    item.addEventListener('click', e => {
      e.preventDefault();
      const page = item.dataset.page;
      navigateToPage(page);
      window.location.hash = '#' + page;
    });
  });

  // Navigate on hash change (browser back/forward)
  window.addEventListener('hashchange', () => {
    const page = window.location.hash.slice(1);
    navigateToPage(page);
  });

  // Navigate to initial hash on load
  const initialPage = window.location.hash.slice(1);
  if (initialPage && VALID_PAGES.has(initialPage)) {
    navigateToPage(initialPage);
  }
}

// ── Poll indicator tooltip ─────────────────────────────────────────────────
// Shared emojivoto state — updated by pollEmojivotoSettings, read by renderPollTooltip.
let lastEmojivotoData = null;

// Updates the data-tooltip on the Polling dot with namespace, mode, and status.
// Called after each successful poll so the info stays current.
function renderPollTooltip(cfg) {
  const indicator = document.querySelector('.poll-indicator');
  if (!indicator) return;

  const lines = [];

  if (cfg.namespace)     lines.push(`Namespace:  ${cfg.namespace}`);
  const modeLabel = cfg.faceMode === 'pubsub' ? 'Pub/Sub' : 'Classic';
  lines.push(`Mode:       ${modeLabel}`);
  if (cfg.faceMode === 'pubsub') {
    const qLabel = cfg.queueBackend === 'redis' ? 'Redis' : 'RabbitMQ';
    lines.push(`Queue:      ${qLabel}`);
    if (cfg.maxDepth) lines.push(`Max depth:  ${cfg.maxDepth}`);
  }
  lines.push(`K8s API:    ${cfg.k8sAvailable ? '✓ connected' : '✗ not available'}`);
  lines.push(`Linkerd:    ${cfg.linkerdMeshed ? 'meshed' : 'not meshed'}`);
  lines.push(`Poll:       every ${POLL_INTERVAL / 1000}s`);

  if (lastEmojivotoData && lastEmojivotoData.enabled) {
    const ev = lastEmojivotoData;
    let evLine = 'Emojivoto:  ';
    if (ev.status === 'ok') {
      evLine += ev.leader ? `✓ ${ev.leader} leading` : '✓ connected';
    } else if (ev.status === 'error') {
      evLine += `✗ ${ev.error || 'error'}`;
    } else {
      evLine += 'connecting…';
    }
    lines.push(evLine);
  }

  indicator.dataset.tooltip = lines.join('\n');
}

// ── Polling ────────────────────────────────────────────────────────────────
async function fetchJSON(url) {
  const r = await fetch(url);
  if (!r.ok) throw new Error(`HTTP ${r.status}`);
  return r.json();
}

async function poll() {
  try {
    const [status, pipeline, config] = await Promise.all([
      fetchJSON('/api/status'),
      fetchJSON('/api/pipeline'),
      fetchJSON('/api/config'),
    ]);
    const prevMode = state.mode;
    state.mode        = status.mode;
    if (prevMode !== status.mode) updateLiveViewModeUI();
    state.queueBackend = status.queueBackend;
    state.lastPipeline = pipeline;

    let controls = null;
    if (status.mode === 'pubsub') {
      controls = await fetchJSON('/api/controls').catch(() => null);
      state.lastControls = controls;
    }

    state.lastConfig    = config;
    lastStatusForHeader = status;
    renderModeBadge(status.mode);
    renderHeaderHealth();
    renderPollTooltip(config);
    renderOverview(status);
    renderPipeline(pipeline, status.mode, controls, status);
    renderControls(controls, status.mode);
    renderConfig(config);

    const chaos = await fetchJSON('/api/chaos').catch(() => null);
    if (chaos) {
      renderChaos(chaos, status.mode);            // Controls page collapse section
      renderFaultInjection(chaos, status.mode);   // dedicated Fault Injection page
    }

    const lu = document.getElementById('last-update');
    if (lu) lu.textContent = new Date().toLocaleTimeString();
  } catch (e) {
    console.warn('poll error', e);
  }
}

// ── Header health mini-bar ─────────────────────────────────────────────────
// Shows compact service status dots on all pages (Overview no longer has its own health grid).
let lastStatusForHeader = null;

function renderHeaderHealth() {
  const bar = document.getElementById('header-health');
  if (!bar || !lastStatusForHeader) return;

  const svcOrder = lastStatusForHeader.mode === 'pubsub'
    ? ['face-publisher','face-subscriber','mysql',
       lastStatusForHeader.queueBackend === 'redis' ? 'redis' : 'rabbitmq',
       'smiley','color','gui']
    : ['face','smiley','color','gui'];

  bar.innerHTML = svcOrder.map(name => {
    const s = lastStatusForHeader.services?.[name];
    if (!s) return '';
    const cls   = s.healthy ? 'ok' : 'bad';
    const title = s.healthy ? `${name} (${s.latencyMs ?? '?'}ms)` : `${name}: ${s.error || 'unhealthy'}`;
    return `<span class="health-chip ${cls}" data-tooltip="${esc(title)}"><span class="h-dot"></span>${name}</span>`;
  }).join('');
}

// ── Mode badge ─────────────────────────────────────────────────────────────
function renderModeBadge(mode) {
  const el = document.getElementById('mode-badge');
  el.textContent = mode === 'pubsub' ? 'Pub/Sub Mode' : 'Classic Mode';
  el.className = 'header-badge ' + (mode === 'pubsub' ? 'badge-pubsub' : 'badge-classic');
  const pipelineNav = document.getElementById('nav-pipeline');
  if (pipelineNav) pipelineNav.style.display = mode === 'pubsub' ? '' : 'none';
  setLiveNote();
}

// ── Overview ───────────────────────────────────────────────────────────────
function renderOverview(status) {
  const classicSection  = document.getElementById('classic-overview-section');
  const pipelineSection = document.getElementById('overview-pipeline-section');

  if (status.mode === 'classic') {
    if (pipelineSection) pipelineSection.style.display = 'none';
    if (classicSection)  classicSection.style.display  = '';
    renderClassicOverview(status);
    return;
  }

  if (classicSection) classicSection.style.display = 'none';
  if (status.mode === 'pubsub' && state.lastPipeline) {
    pipelineSection.style.display = '';
    renderPubSubArchitecture(status);
    const pl = state.lastPipeline;
    const m = pl.mysql, q = pl.queue;
    const ctrl = state.lastControls;

    // Diagram — same logic as renderPipeline
    const pubPaused = ctrl?.publisher?.paused ?? false;
    const subPaused = ctrl?.subscriber?.paused ?? false;
    const pubArrow = pubPaused ? 'blocked' : 'flowing';
    const subArrow = subPaused ? 'blocked' : 'flowing';

    let pubCount, pubLabel;
    if (pubPaused) { pubCount = '⏸'; pubLabel = 'paused'; }
    else if (ctrl?.publisher?.available) {
      const ms = ctrl.publisher.publishIntervalMs;
      pubCount = msToRate(ms); pubLabel = ms === 0 ? 'no delay' : 'msg/sec';
    } else { pubCount = '∞'; pubLabel = 'publishing'; }

    // MySQL shows 'queued' — messages written AND pushed to queue, awaiting subscriber.
    // 'pending' (~0 normally) is only non-zero during crash-recovery window.
    const ovMysqlQueued = m.available ? fmt(m.queued ?? m.pending) : 'N/A';
    const subCount = subPaused ? '⏸' : (m.available ? fmt(m.acknowledged) : 'N/A');
    const subLabel = subPaused ? 'paused' : (m.available ? 'delivered' : '');

    const qTT = q.backend === 'redis' ? TT.redis : TT.rabbitmq;
    const ovPending = m.available ? (m.pending ?? 0) : 0;
    const ovPubLatMs  = status?.services?.['face-publisher']?.latencyMs;
    const ovSubLatMs  = status?.services?.['face-subscriber']?.latencyMs;
    const ovPubLatStr = ovPubLatMs != null ? ` · ${ovPubLatMs}ms` : '';
    const ovSubSec    = ovSubLatMs != null ? `${ovSubLatMs}ms latency` : '';
    const ovMysqlSec = ovPending > 0 ? `⚠ ${fmt(ovPending)} pending` : `${fmt(m.queued??0)} in queue${ovPubLatStr}`;
    document.getElementById('overview-pipeline-diagram').innerHTML = `
      ${pipeNode('📤', 'Publisher', pubCount, pubLabel, false, pubPaused, TT.publisher)}
      ${pipeArrow('write-ahead', pubArrow)}
      ${pipeNode('🗄️', 'MySQL', ovMysqlQueued, 'queued', !m.available, false, TT.mysql, '', ovMysqlSec, ovPending > 0)}
      ${pipeArrow('push', pubArrow)}
      ${pipeNode(q.backend==='redis'?'⚡':'🐰', q.backend, q.available ? fmt(q.depth) : 'N/A', `of ${fmt(q.maxDepth)}`, !q.available, false, qTT)}
      ${pipeArrow('pop + ack', subArrow)}
      ${pipeNode('📥', 'Subscriber', subCount, subLabel, !m.available && !subPaused, subPaused, TT.subscriber, '', ovSubSec)}
      ${pipeArrow('deliver', subArrow)}
      ${pipeNode('🖥️', 'GUI', '∞', 'consuming', false, false, TT.gui)}
    `;

    const ovStranded = (m.available && q.available && m.queued != null)
      ? Math.max(0, (m.queued ?? 0) - q.depth) : null;
    const ovNetRate = (q.available && q.readyRate != null && q.deliverRate != null)
      ? q.readyRate - q.deliverRate : null;
    const netCls = ovNetRate === null ? ''
      : Math.abs(ovNetRate) < 3 ? 'balanced'
      : ovNetRate > 0 ? 'warn' : 'draining';
    const netLabel = ovNetRate === null ? '–'
      : Math.abs(ovNetRate) < 3 ? `≈0 balanced`
      : ovNetRate > 0 ? `+${ovNetRate.toFixed(1)} filling` : `${ovNetRate.toFixed(1)} draining`;
    document.getElementById('overview-metrics').innerHTML = `
      ${metricCard(m.available ? fmt(m.queued ?? m.pending) : '–', 'DB Queued',   '',     TT.db_queued)}
      ${metricCard(m.available ? fmt(m.acknowledged) : '–',  'Delivered',  '',            TT.delivered)}
      ${metricCard(q.available ? fmt(q.depth) : '–', `${q.backend} Depth`, q.depth > q.maxDepth * 0.8 ? 'warn' : '', TT.depth)}
      ${metricCard(ovStranded !== null ? fmt(ovStranded) : '–', 'Dropped', ovStranded > 100 ? 'warn' : '', TT.dropped)}
      ${metricCard(q.available && q.readyRate   ? q.readyRate.toFixed(1)+' msg/s'   : '–', 'Pub Rate', '', TT.pub_rate)}
      ${metricCard(q.available && q.deliverRate ? q.deliverRate.toFixed(1)+' msg/s' : '–', 'Del Rate', '', TT.del_rate)}
      ${metricCard(netLabel, 'Net Rate', netCls, 'Pub Rate minus Del Rate. ≈0 means the queue is stable — publish rate matches GUI consumption. Positive means the queue is filling (reduce pub rate). Negative means it\'s draining (queue will empty).')}
    `;
  } else {
    pipelineSection.style.display = 'none';
  }
}

// ── Pub/Sub architecture diagram ───────────────────────────────────────────
// Static topology view: Smiley + Color feed into Publisher, which writes to
// MySQL → Queue → Subscriber → GUI. Latency and health from /api/status.
function renderPubSubArchitecture(status) {
  const arch = document.getElementById('pubsub-arch-diagram');
  if (!arch) return;

  const svcs = status.services || {};
  const ok  = n => svcs[n]?.healthy ?? false;
  const lat = n => svcs[n]?.latencyMs != null ? svcs[n].latencyMs + 'ms' : '–';
  // For infrastructure nodes (MySQL, queue) that don't have response times
  // from the user's perspective but do have a ping latency — show ms if available,
  // fall back to ✓ for healthy / – for unhealthy.
  const latOrCheck = n => {
    const s = svcs[n];
    if (!s) return '–';
    if (!s.healthy) return '–';
    return s.latencyMs != null ? s.latencyMs + 'ms' : '✓';
  };

  const nb = (icon, name, val, healthy) => `
    <div class="chain-node-box ${healthy ? 'ok' : 'bad'}">
      <span class="cn-icon">${icon}</span>
      <div class="cn-name">${esc(name)}</div>
      <div class="cn-value">${esc(val)}</div>
    </div>`;

  const arrow = label => `
    <div class="chain-arrow-h psarch-arrow">
      <div class="ah-arrow-wrap">
        <div class="ah-line"></div>
        <div class="ah-head"></div>
      </div>
      <div class="ah-label">${esc(label)}</div>
    </div>`;

  const qIcon = status.queueBackend === 'redis' ? '⚡' : '🐰';
  const qName = status.queueBackend === 'redis' ? 'Redis' : 'RabbitMQ';

  const forkHeight = 80;

  arch.innerHTML = `
    <div class="chain-fork reversed">
      <div class="fork-branches">
        <div class="fork-branch">
          <div class="fork-hbar"></div>
          <span class="fork-label">HTTP</span>
          ${nb('😃', 'Smiley', lat('smiley'), ok('smiley'))}
        </div>
        <div class="fork-branch">
          <div class="fork-hbar"></div>
          <span class="fork-label">gRPC</span>
          ${nb('🎨', 'Color', lat('color'), ok('color'))}
        </div>
      </div>
      <div class="fork-vbar" style="height:${forkHeight}px"></div>
    </div>
    ${nb('📤', 'Publisher', lat('face-publisher'), ok('face-publisher'))}
    ${arrow('write')}
    ${nb('🗄️', 'MySQL', latOrCheck('mysql'), ok('mysql'))}
    ${arrow('push')}
    ${nb(qIcon, qName, latOrCheck(status.queueBackend || 'rabbitmq'), ok(status.queueBackend || 'rabbitmq'))}
    ${arrow('pop')}
    ${nb('📥', 'Subscriber', lat('face-subscriber'), ok('face-subscriber'))}
    ${arrow('deliver')}
    ${nb('🖥️', 'GUI', lat('gui'), ok('gui'))}`;
}

// ── Classic mode overview ──────────────────────────────────────────────────
function renderClassicOverview(status) {
  const svcs = status.services || {};

  // ── Service health cards ───────────────────────────────────────────────
  const grid = document.getElementById('classic-services-grid');
  if (grid) {
    const order = ['face', 'smiley', 'color', 'gui'];
    grid.innerHTML = order.map(name => {
      const s = svcs[name];
      if (!s) return '';
      const cls    = s.healthy ? 'healthy' : 'unhealthy';
      const dotCls = s.healthy ? 'dot-green' : 'dot-red';
      const latency = s.latencyMs ? `<div class="service-latency">${s.latencyMs}ms</div>` : '';
      const errHtml = s.error
        ? `<div class="service-latency error-text">${esc(s.error)}</div>` : '';
      return `
        <div class="service-card ${cls}">
          <div class="service-name">${esc(name)}</div>
          <div class="service-status"><div class="dot ${dotCls}"></div>${s.healthy ? 'Healthy' : 'Unhealthy'}</div>
          ${latency}${errHtml}
        </div>`;
    }).join('');
  }

  // ── Request chain diagram (fork: Face calls Smiley + Color in parallel) ──
  const chain = document.getElementById('classic-chain-diagram');
  if (chain) {
    const svc     = name => svcs[name] || {};
    const ok      = name => svc(name).healthy;
    const latency = name => svc(name).latencyMs ? svc(name).latencyMs + 'ms' : '–';

    // Helper: a single node box
    const nodeBox = (icon, name, val, healthy) => `
      <div class="chain-node-box ${healthy ? 'ok' : 'bad'}">
        <span class="cn-icon">${icon}</span>
        <div class="cn-name">${esc(name)}</div>
        <div class="cn-value">${esc(val)}</div>
      </div>`;

    // Helper: horizontal arrow
    const arrowH = flowing => `
      <div class="chain-arrow-h">
        <div class="ah-line" style="${flowing ? '' : 'background:linear-gradient(90deg,var(--red),rgba(248,113,113,.15))'}"></div>
        <div class="ah-head" style="${flowing ? '' : 'border-left-color:var(--red)'}"></div>
      </div>`;

    // Determine fork vbar height based on gap between branches (must match CSS gap: 16px + node height ≈ 60px)
    const forkHeight = 76;

    chain.innerHTML = `
      ${nodeBox('🌐', 'GUI', latency('gui'), ok('gui'))}
      ${arrowH(ok('gui'))}
      ${nodeBox('😀', 'Face', latency('face'), ok('face'))}

      <div class="chain-fork">
        <div class="fork-vbar" style="height:${forkHeight}px"></div>
        <div class="fork-branches">
          <div class="fork-branch">
            <div class="fork-hbar"></div>
            <span class="fork-label">HTTP</span>
            ${nodeBox('😃', 'Smiley', latency('smiley'), ok('smiley'))}
          </div>
          <div class="fork-branch">
            <div class="fork-hbar"></div>
            <span class="fork-label">gRPC</span>
            ${nodeBox('🎨', 'Color', latency('color'), ok('color'))}
          </div>
        </div>
      </div>
    `;
  }
}

// ── Pipeline ───────────────────────────────────────────────────────────────
function renderPipeline(pipeline, mode, controls, status = null) {
  const unavail = document.getElementById('pipeline-unavailable');
  if (mode !== 'pubsub') {
    unavail.style.display = '';
    document.getElementById('pipeline-diagram').innerHTML = '';
    document.getElementById('pipeline-metrics').innerHTML = '';
    return;
  }
  unavail.style.display = 'none';

  const m = pipeline.mysql, q = pipeline.queue;
  const pubPaused = controls?.publisher?.paused ?? false;
  const subPaused = controls?.subscriber?.paused ?? false;
  const pubArrow = pubPaused ? 'blocked' : 'flowing';
  const subArrow = subPaused ? 'blocked' : 'flowing';

  let pubCount, pubLabel;
  if (pubPaused) {
    pubCount = '⏸'; pubLabel = 'paused';
  } else if (controls?.publisher?.available) {
    const ms = controls.publisher.publishIntervalMs;
    pubCount = msToRate(ms);
    pubLabel = ms === 0 ? 'no delay' : 'msg/sec';
  } else {
    pubCount = '∞'; pubLabel = 'publishing';
  }

  // MySQL node shows 'queued': messages pushed to queue, waiting for subscriber.
  // 'pending' should be ~0; non-zero means queue backend unreachable.
  const mysqlQueued = m.available ? fmt(m.queued ?? m.pending) : 'N/A';
  const subCount = subPaused ? '⏸' : (m.available ? fmt(m.acknowledged) : 'N/A');
  const subLabel = subPaused ? 'paused' : (m.available ? 'delivered' : '');

  const qTipText = q.backend === 'redis' ? TT.redis : TT.rabbitmq;
  const pendingNow2 = m.available ? (m.pending ?? 0) : 0;

  // Latency from health checks: show as secondary on publisher/subscriber nodes
  const pubLatMs = status?.services?.['face-publisher']?.latencyMs;
  const subLatMs = status?.services?.['face-subscriber']?.latencyMs;
  const pubLatStr  = pubLatMs != null ? ` · ${pubLatMs}ms` : '';
  const subSecondary = subLatMs != null ? `${subLatMs}ms latency` : '';

  const mysqlSecondary = pendingNow2 > 0
    ? `⚠ ${fmt(pendingNow2)} pending`
    : `${fmt(m.queued??0)} in queue${pubLatStr}`;
  document.getElementById('pipeline-diagram').innerHTML = `
    ${pipeNode('📤', 'Publisher', pubCount, pubLabel, false, pubPaused, TT.publisher)}
    ${pipeArrow('write-ahead', pubArrow)}
    ${pipeNode('🗄️', 'MySQL', mysqlQueued, 'queued', !m.available, false, TT.mysql, '', mysqlSecondary, pendingNow2 > 0)}
    ${pipeArrow('push', pubArrow)}
    ${pipeNode(q.backend==='redis'?'⚡':'🐰', q.backend, q.available ? fmt(q.depth) : 'N/A', `of ${fmt(q.maxDepth)}`, !q.available, false, qTipText)}
    ${pipeArrow('pop + ack', subArrow)}
    ${pipeNode('📥', 'Subscriber', subCount, subLabel, !m.available && !subPaused, subPaused, TT.subscriber, '', subSecondary)}
    ${pipeArrow('deliver', subArrow)}
    ${pipeNode('🖥️', 'GUI', '∞', 'consuming', false, false, TT.gui)}
  `;

  const stranded = (m.available && q.available && m.queued != null)
    ? Math.max(0, (m.queued ?? 0) - q.depth) : null;

  const netRate = (q.available && q.readyRate != null && q.deliverRate != null)
    ? q.readyRate - q.deliverRate : null;
  const netCls = netRate === null ? ''
    : Math.abs(netRate) < 3 ? 'balanced'
    : netRate > 0 ? 'warn' : 'draining';
  const netLabel = netRate === null ? '–'
    : Math.abs(netRate) < 3 ? `≈0 balanced`
    : netRate > 0 ? `+${netRate.toFixed(1)} filling` : `${netRate.toFixed(1)} draining`;

  document.getElementById('pipeline-metrics').innerHTML = `
    ${metricCard(m.available ? fmt(m.queued ?? m.pending) : '–', 'DB Queued',  '', TT.db_queued)}
    ${metricCard(m.available ? fmt(m.acknowledged) : '–', 'Delivered', '',         TT.delivered)}
    ${metricCard(q.available ? fmt(q.depth) : '–', `${q.backend} Depth`, q.depth > q.maxDepth*0.8 ? 'warn' : '', TT.depth)}
    ${metricCard(stranded !== null ? fmt(stranded) : '–', 'Dropped', stranded > 100 ? 'warn' : '', TT.dropped)}
    ${metricCard(q.available && q.readyRate   ? q.readyRate.toFixed(1)+' msg/s'   : '–', 'Pub Rate', '', TT.pub_rate)}
    ${metricCard(q.available && q.deliverRate ? q.deliverRate.toFixed(1)+' msg/s' : '–', 'Del Rate', '', TT.del_rate)}
    ${metricCard(netLabel, 'Net Rate', netCls, 'Pub Rate minus Del Rate. ≈0 means queue is stable — publish rate matches GUI consumption. Positive means filling (reduce rate). Negative means draining (queue will empty).')}
  `;

  // Warning 1: pending persistently high → queue backend unreachable.
  // At high publish rates (100–300 msg/s) it's normal to see 1–20 pending rows during
  // Pending warning removed — was too noisy (false positives during queue full,
  // backend switches, chaos injection). The MySQL node's secondary line already
  // shows "⚠ N pending" inline when applicable.

  // Warning 2: stranded rows — queued in MySQL but missing from the queue backend.
  // Only meaningful when the queue is NOT full (if it's full, rows are just waiting
  // for the queue to drain before they can be pushed — that's normal backpressure,
  // not eviction, and re-warm would be a no-op or harmful).
  const droppedThreshold = q.maxDepth > 0
    ? Math.max(1000, Math.round(q.maxDepth * 0.2))
    : 1000;
  const queueFull  = q.available && q.maxDepth > 0 && q.depth >= q.maxDepth * 0.95;
  const warn = document.getElementById('dropped-warning');
  if (stranded !== null && stranded > droppedThreshold && !queueFull) {
    warn.classList.add('visible');
    document.getElementById('dropped-count').textContent = fmt(stranded);
  } else {
    warn.classList.remove('visible');
  }

  if (pipeline.history?.length) { state.history = pipeline.history; drawCharts(); }
}

// ── Flow controls ──────────────────────────────────────────────────────────
function renderControls(controls, mode) {
  const section = document.getElementById('flow-controls-section');
  if (mode !== 'pubsub') { section.style.display = 'none'; return; }
  section.style.display = '';

  const cards = document.getElementById('control-cards');
  cards.innerHTML = '';

  const pubCtrl = controls?.publisher;
  cards.appendChild(buildControlCard({
    id:           'ctrl-publisher',
    icon:         '📤',
    label:        'Publisher',
    ctrl:         pubCtrl,
    meta:         pubCtrl?.available ? [{ icon: '⚙', text: `${pubCtrl.publishConcurrency} loop${pubCtrl.publishConcurrency !== 1 ? 's' : ''} per pod` }] : [],
    onToggle:     paused => togglePause('publisher', paused),
    currentRateMs: pubCtrl?.publishIntervalMs ?? null,
    onSetRate:    pubCtrl?.available ? ms => setPublishRate(ms) : null,
  }));

  cards.appendChild(buildControlCard({
    id:       'ctrl-subscriber',
    icon:     '📥',
    label:    'Subscriber',
    ctrl:     controls?.subscriber,
    meta:     [],
    onToggle: paused => togglePause('subscriber', paused),
  }));
}

function buildControlCard({ id, icon, label, ctrl, meta, onToggle, currentRateMs, onSetRate }) {
  const available = ctrl?.available ?? false;
  const paused    = ctrl?.paused    ?? false;

  let stateClass, badgeClass, badgeText;
  if (!available)    { stateClass = 'state-unavailable'; badgeClass = 'unknown';  badgeText = 'Unavailable'; }
  else if (paused)   { stateClass = 'state-paused';      badgeClass = 'paused';   badgeText = 'Paused'; }
  else               { stateClass = 'state-running';     badgeClass = 'running';  badgeText = 'Running'; }

  const card = document.createElement('div');
  card.className = `control-card ${stateClass}`;
  card.id = id;

  const metaHtml = meta.length
    ? `<div class="ctrl-meta">${meta.map(m => `<span>${m.icon} ${esc(m.text)}</span>`).join('')}</div>`
    : '';

  const podCountHtml = ctrl?.podCount > 0
    ? `<span style="margin-left:8px;font-size:11px;color:var(--text-muted)">· ${ctrl.podCount} pod${ctrl.podCount !== 1 ? 's' : ''}</span>`
    : '';

  card.innerHTML = `
    <div class="ctrl-header">
      <span class="ctrl-title">${icon} ${esc(label)}${podCountHtml}</span>
      <span class="ctrl-badge ${badgeClass}">${esc(badgeText)}</span>
    </div>
    ${metaHtml}
    <button class="ctrl-btn ${paused ? 'resume' : 'pause'}" ${!available ? 'disabled' : ''}>${paused ? '▶ Resume' : '⏸ Pause'}</button>
  `;

  card.querySelector('.ctrl-btn').addEventListener('click', () => onToggle(!paused));

  // Pod breakdown — inserted between meta and rate control / button
  if (ctrl?.pods?.length > 0) {
    const podSection = buildPodList(ctrl.pods, ctrl.publishConcurrency);
    const insertBefore = card.querySelector('.ctrl-btn');
    card.insertBefore(podSection, insertBefore);
  }

  if (onSetRate && currentRateMs !== null) {
    card.insertBefore(buildRateSlider(id, currentRateMs, available, onSetRate),
      card.querySelector('.ctrl-btn'));
  }

  return card;
}

function buildPodList(pods, concurrencyPerPod) {
  const wrap = document.createElement('div');
  wrap.className = 'ctrl-pod-list';

  for (const p of pods) {
    const row = document.createElement('div');
    row.className = 'ctrl-pod-row';

    const dotCls = !p.available ? 'error' : p.paused ? 'paused' : 'running';
    const statusText = !p.available ? 'unavailable' : p.paused ? 'paused' : 'running';

    // Prefer pod name from K8s API; fall back to last-2-octets of IP
    const displayName = p.podName
      ? shortPodName(p.podName)
      : (p.podIP ? p.podIP.split('.').slice(-2).join('.') : '?');
    // Multi-line tooltip — pre-line CSS renders \n as real line breaks
    const titleText = podTooltipText({ name: p.podName, ip: p.podIP, zone: p.zone, region: p.region, node: p.node });
    const zoneBadge = p.zone ? `<span class="zone-badge" data-tooltip="Zone: ${esc(p.zone)}">${esc(p.zone)}</span>` : '';

    // Rate for publisher pods
    const rateHtml = p.publishIntervalMs > 0
      ? `<span class="ctrl-pod-rate">${msToRate(p.publishIntervalMs)}</span>`
      : '';

    const errHtml = p.error
      ? `<span class="ctrl-pod-err" title="${esc(p.error)}">${esc(p.error)}</span>`
      : `<span class="ctrl-pod-status">${statusText}</span>`;

    row.innerHTML = `
      <span class="ctrl-pod-dot ${dotCls}"></span>
      <span class="ctrl-pod-ip" data-tooltip="${esc(titleText)}">${esc(displayName)}</span>
      ${zoneBadge}
      ${rateHtml}
      ${errHtml}
    `;
    wrap.appendChild(row);
  }

  // Total throughput row for publisher pods
  const publisherPods = pods.filter(p => p.available && p.publishIntervalMs > 0);
  if (publisherPods.length > 0) {
    const totalMsgPerSec = publisherPods.reduce((sum, p) => {
      const ratePerLoop = p.publishIntervalMs === 0 ? 0 : 1000 / p.publishIntervalMs;
      return sum + ratePerLoop * (p.publishConcurrency || concurrencyPerPod || 1);
    }, 0);

    const totalRow = document.createElement('div');
    totalRow.className = 'ctrl-total-rate';
    totalRow.innerHTML = `
      <span data-tooltip="Theoretical target (interval × loops × pods). Actual throughput also spends time per message calling smiley/color and writing to MySQL + the queue — the Pub Rate card shows the measured rate.">Target across ${publisherPods.length} pod${publisherPods.length !== 1 ? 's' : ''}</span>
      <strong>~${Math.round(totalMsgPerSec)} msg/s</strong>
    `;
    wrap.appendChild(totalRow);
  }

  return wrap;
}

function buildRateSlider(cardId, currentMs, enabled, onSetRate) {
  const section = document.createElement('div');
  section.className = 'ctrl-rate-section';

  const sliderVal = msToSlider(currentMs);
  const isFlood   = currentMs === 0;

  section.innerHTML = `
    <div class="ctrl-rate-header">
      <span class="ctrl-rate-label">Publish Rate (msg/s)</span>
      <span class="ctrl-rate-value ${isFlood ? 'flood' : ''}" id="${cardId}-rate-val">${esc(msToRate(currentMs))}</span>
    </div>
    <input type="range" class="ctrl-rate-slider" id="${cardId}-rate-slider"
           min="0" max="100" step="1" value="${sliderVal}"
           style="--fill:${sliderVal}%"
           ${!enabled ? 'disabled' : ''}>
    <div class="ctrl-rate-ticks">
      ${RATE_TICKS.map((t, i) => {
        const pct  = msToSlider(t.ms);
        // prevent first/last label from being clipped by the container edges
        const xform = i === 0
          ? 'translateX(0)'
          : i === RATE_TICKS.length - 1
            ? 'translateX(-100%)'
            : 'translateX(-50%)';
        return `<button class="rate-tick ${t.ms === currentMs ? 'active' : ''}"
                        data-ms="${t.ms}" style="left:${pct}%;transform:${xform}"
                        ${!enabled ? 'disabled' : ''}>${esc(msToRate(t.ms))}</button>`;
      }).join('')}
    </div>
  `;

  let debounceTimer = null;

  const slider  = section.querySelector(`#${cardId}-rate-slider`);
  const valDisp = section.querySelector(`#${cardId}-rate-val`);

  function applyValue(ms) {
    const sv    = msToSlider(ms);
    const flood = ms === 0;
    slider.value = sv;
    slider.style.setProperty('--fill', sv + '%');
    slider.dataset.flood = flood ? 'true' : 'false';
    valDisp.textContent  = msToRate(ms);
    valDisp.className    = `ctrl-rate-value ${flood ? 'flood' : ''}`;
    section.querySelectorAll('.rate-tick').forEach((b, i) => {
      b.classList.toggle('active', Number(b.dataset.ms) === ms);
    });
  }

  slider.addEventListener('input', () => {
    const ms = sliderToMs(Number(slider.value));
    applyValue(ms);
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => onSetRate(ms), 300);
  });

  section.querySelectorAll('.rate-tick').forEach(btn => {
    btn.addEventListener('click', () => {
      const ms = Number(btn.dataset.ms);
      applyValue(ms);
      clearTimeout(debounceTimer);
      onSetRate(ms);
    });
  });

  return section;
}

let rateDebounce = null;
async function setPublishRate(ms) {
  try {
    const r = await fetch('/api/controls/publisher', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ publishIntervalMs: ms }),
    });
    if (r.ok) showToast(`⚡ Rate → ${msToRate(ms)}`, 'success');
    else      showToast(`Rate change failed: HTTP ${r.status}`, 'error');
  } catch (e) {
    showToast('Publisher unreachable', 'error');
  }
  // Don't call poll() here — the slider already shows the right value optimistically.
  // The next scheduled poll will sync any discrepancy.
}

async function togglePause(service, paused) {
  setControlsDisabled(true);
  try {
    const r = await fetch(`/api/controls/${service}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ paused }),
    });
    if (r.ok) showToast(`${paused ? '⏸' : '▶'} ${paused ? 'Paused' : 'Resumed'} ${service}`, paused ? 'warn' : 'success');
    else      showToast(`Failed: HTTP ${r.status}`, 'error');
  } catch (e) {
    showToast(`${service} unreachable`, 'error');
  } finally {
    setControlsDisabled(false);
    poll();
  }
}

async function runScenario(scenario) {
  const SCENARIOS = {
    drain:  { publisher: true,  subscriber: false, label: '▼ Draining queue…', type: 'warn' },
    fill:   { publisher: false, subscriber: true,  label: '▲ Filling queue…',  type: 'warn' },
    freeze: { publisher: true,  subscriber: true,  label: '⏸ Frozen.',          type: 'warn' },
    resume: { publisher: false, subscriber: false,  label: '▶ Resumed.',        type: 'success' },
  };
  const s = SCENARIOS[scenario];
  if (!s) return;
  setScenarioDisabled(true);
  try {
    await Promise.all([
      fetch('/api/controls/publisher',  { method: 'PUT', headers: {'Content-Type':'application/json'}, body: JSON.stringify({paused: s.publisher}) }),
      fetch('/api/controls/subscriber', { method: 'PUT', headers: {'Content-Type':'application/json'}, body: JSON.stringify({paused: s.subscriber}) }),
    ]);
    showToast(s.label, s.type);
  } catch (e) {
    showToast('Scenario failed: ' + e.message, 'error');
  } finally {
    setScenarioDisabled(false);
    poll();
  }
}

function setControlsDisabled(d)  { document.querySelectorAll('.ctrl-btn').forEach(b => { b.disabled = d; }); }
function setScenarioDisabled(d)  { document.querySelectorAll('.scenario-btn').forEach(b => { b.disabled = d; }); }

function initScenarioButtons() {
  document.querySelectorAll('.scenario-btn').forEach(btn =>
    btn.addEventListener('click', () => runScenario(btn.dataset.scenario)));
}

// ── Charts ─────────────────────────────────────────────────────────────────
function drawCharts() {
  const h = state.history;
  if (!h || h.length < 2) return;
  const depthSeries = [
    { data: h.map(p => p.queueDepth),   color: '#53d8fb' },
    { data: h.map(p => p.queued ?? p.pending), color: '#fbbf24' },
  ];

  // Delivery rate (msgs/s) — derivative of the cumulative acknowledged counter.
  // The raw counter only ever creeps upward, which reads as a meaningless ramp;
  // the per-second rate actually shows consumption starting, stopping, and
  // draining. Negative deltas (DB purge/reset) clamp to 0, and a light
  // 3-sample average smooths per-poll jitter.
  const raw = [];
  for (let i = 1; i < h.length; i++) {
    const dt = (h[i].ts - h[i - 1].ts) / 1000;
    const d  = h[i].acknowledged - h[i - 1].acknowledged;
    raw.push(dt > 0 && d >= 0 ? d / dt : 0);
  }
  const rateData = [null]; // first history point has no delta
  for (let i = 0; i < raw.length; i++) {
    const win = raw.slice(Math.max(0, i - 2), i + 1);
    rateData.push(win.reduce((s, v) => s + v, 0) / win.length);
  }
  const rateSeries = [{ data: rateData, color: '#4ade80' }];

  // Draw to both the Pipeline page and the Overview page canvases
  drawSparkline('chart-depth',          depthSeries);
  drawSparkline('chart-acked',          rateSeries);
  drawSparkline('overview-chart-depth', depthSeries);
  drawSparkline('overview-chart-acked', rateSeries);

  // Live current values in the chart-card headers
  const setVal = (id, v) => { const el = document.getElementById(id); if (el) el.textContent = v; };
  const depthNow = fmt(h.at(-1).queueDepth ?? 0);
  const rateNow  = `${(rateData.at(-1) ?? 0).toFixed(1)}/s`;
  setVal('depth-now', depthNow);
  setVal('overview-depth-now', depthNow);
  setVal('rate-now', rateNow);
  setVal('overview-rate-now', rateNow);
}

function drawSparkline(id, series) {
  const canvas = document.getElementById(id);
  if (!canvas) return;
  const dpr = window.devicePixelRatio || 1;
  const W = canvas.offsetWidth || 800, H = canvas.offsetHeight || 80;
  canvas.width = W * dpr; canvas.height = H * dpr;
  const ctx = canvas.getContext('2d');
  ctx.scale(dpr, dpr);

  const isDark = document.documentElement.dataset.theme !== 'light';
  ctx.fillStyle = isDark ? '#1c1c2e' : '#ffffff';
  ctx.fillRect(0, 0, W, H);

  const all = series.flatMap(s => s.data).filter(v => v != null);
  if (!all.length) return;
  const maxV = Math.max(...all, 1);
  const pad = 6, n = series[0].data.length;
  const xPos = i => pad + (i / (n - 1)) * (W - pad * 2);
  const yPos = v => H - pad - (v / maxV) * (H - pad * 2);

  ctx.strokeStyle = isDark ? '#2a2a45' : '#e2e8f0';
  ctx.lineWidth = 1;
  for (let i = 1; i < 4; i++) {
    const y = pad + (i / 4) * (H - pad * 2);
    ctx.beginPath(); ctx.moveTo(pad, y); ctx.lineTo(W - pad, y); ctx.stroke();
  }

  for (const s of series) {
    ctx.beginPath(); ctx.strokeStyle = s.color; ctx.lineWidth = 1.5;
    let started = false;
    for (let i = 0; i < s.data.length; i++) {
      const v = s.data[i]; if (v == null) continue;
      if (!started) { ctx.moveTo(xPos(i), yPos(v)); started = true; }
      else ctx.lineTo(xPos(i), yPos(v));
    }
    ctx.stroke();
    const first = s.data.findIndex(v => v != null);
    const last  = s.data.length - 1 - [...s.data].reverse().findIndex(v => v != null);
    if (first >= 0 && last > first) {
      ctx.beginPath();
      for (let i = first; i <= last; i++) {
        const v = s.data[i]; if (v == null) continue;
        if (i === first) ctx.moveTo(xPos(i), yPos(v)); else ctx.lineTo(xPos(i), yPos(v));
      }
      ctx.lineTo(xPos(last), H - pad); ctx.lineTo(xPos(first), H - pad);
      ctx.closePath(); ctx.fillStyle = s.color + '22'; ctx.fill();
    }
  }
}

// ── Config / Settings ──────────────────────────────────────────────────────
// Cache original (env-var) values so we can detect overrides and support reset
let _cfgOriginal = null;

function renderConfig(cfg) {
  if (!_cfgOriginal) _cfgOriginal = cfg; // save first-load values as defaults

  // Show sign-out button only when auth is active
  const signOutBtn = document.getElementById('sign-out-btn');
  if (signOutBtn) signOutBtn.style.display = cfg.authEnabled ? '' : 'none';

  updateModeToggleButtons(cfg.faceMode);

  const rows = [
    ['Face Mode', cfg.faceMode], ['Smiley Service', cfg.smileyURL],
    ['Color Service', cfg.colorURL], ['GUI Service', cfg.guiURL], ['Face Service', cfg.faceURL],
  ];
  if (cfg.faceMode === 'pubsub') {
    // Queue/publisher/subscriber rows only exist in the pub/sub pipeline
    rows.push(
      ['Queue Backend', cfg.queueBackend], ['Max Queue Depth', cfg.maxDepth],
      ['Publisher', cfg.publisherURL], ['Subscriber', cfg.subscriberURL],
    );
  }
  document.getElementById('config-tbody').innerHTML = rows
    .map(([k, v]) => `<tr><td>${esc(k)}</td><td>${esc(String(v ?? '–'))}</td></tr>`)
    .join('');

  // Show maintenance section only in pubsub mode
  const maintSection = document.getElementById('settings-maintenance');
  if (maintSection) maintSection.style.display = cfg.faceMode === 'pubsub' ? '' : 'none';

  const qTitle = document.getElementById('maint-queue-title');
  if (qTitle) qTitle.textContent = cfg.queueBackend === 'redis' ? '⚡ Redis Queue' : '🐰 RabbitMQ Queue';
  const rmqWrap = document.getElementById('maint-rmq-cmd-wrap');
  if (rmqWrap) rmqWrap.style.display = cfg.queueBackend === 'rabbitmq' ? '' : 'none';

  // Build editable endpoint fields (only on first render to avoid losing focus)
  const grid = document.getElementById('endpoint-grid');
  if (grid && grid.children.length === 0) buildEndpointGrid(cfg);
  else updateEndpointOverrideBadges(cfg);
}

// The editable endpoints — label, config key, current value
function endpointFields(cfg) {
  const fields = [
    { label: 'Smiley',      key: 'smileyURL',     val: cfg.smileyURL },
    { label: 'Color',       key: 'colorURL',       val: cfg.colorURL },
    { label: 'Face / GUI',  key: 'faceURL',        val: cfg.faceURL },
    { label: 'GUI',         key: 'guiURL',         val: cfg.guiURL },
  ];
  if (cfg.faceMode === 'pubsub') {
    fields.push(
      { label: 'Publisher',  key: 'publisherURL',  val: cfg.publisherURL },
      { label: 'Subscriber', key: 'subscriberURL', val: cfg.subscriberURL },
    );
  }
  return fields;
}

function buildEndpointGrid(cfg) {
  const grid = document.getElementById('endpoint-grid');
  if (!grid) return;
  grid.innerHTML = '';

  endpointFields(cfg).forEach(({ label, key, val }) => {
    const orig = _cfgOriginal?.[key] ?? val;
    const overridden = val !== orig;

    const row = document.createElement('div');
    row.className = 'endpoint-row';
    row.innerHTML = `
      <span class="endpoint-label" id="ep-label-${key}">
        ${esc(label)}
        ${overridden ? '<span class="ep-overridden">modified</span>' : ''}
      </span>
      <input class="endpoint-input ${overridden ? 'modified' : ''}"
             id="ep-input-${key}" data-key="${key}" data-orig="${esc(orig)}"
             type="text" value="${esc(val)}" spellcheck="false">
    `;
    grid.appendChild(row);

    row.querySelector('input').addEventListener('input', e => {
      e.target.classList.toggle('modified', e.target.value !== e.target.dataset.orig);
    });
  });
}

function updateEndpointOverrideBadges(cfg) {
  endpointFields(cfg).forEach(({ key, val }) => {
    const input = document.getElementById(`ep-input-${key}`);
    const label = document.getElementById(`ep-label-${key}`);
    if (!input || !label) return;
    const orig = input.dataset.orig || val;
    const isModified = val !== orig;
    const badge = label.querySelector('.ep-overridden');
    if (isModified && !badge) {
      label.insertAdjacentHTML('beforeend', '<span class="ep-overridden">modified</span>');
    } else if (!isModified && badge) {
      badge.remove();
    }
  });
}

function initEndpointEditor() {
  document.getElementById('endpoints-save')?.addEventListener('click', async () => {
    const body = {};
    document.querySelectorAll('.endpoint-input').forEach(input => {
      if (input.value.trim()) body[input.dataset.key] = input.value.trim();
    });
    const status = document.getElementById('endpoints-status');
    try {
      const r = await fetch('/api/config', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (r.ok) {
        status.textContent = '✓ Saved — takes effect immediately';
        status.style.color = 'var(--green)';
        showToast('✓ Service endpoints updated', 'success');
        poll(); // refresh to confirm the new URLs are reflected
      } else {
        status.textContent = 'Save failed';
        status.style.color = 'var(--red)';
      }
    } catch (e) {
      status.textContent = e.message;
      status.style.color = 'var(--red)';
    }
    setTimeout(() => { status.textContent = ''; }, 4000);
  });

  document.getElementById('endpoints-reset')?.addEventListener('click', () => {
    document.querySelectorAll('.endpoint-input').forEach(input => {
      input.value = input.dataset.orig || '';
      input.classList.remove('modified');
    });
    // Send empty strings to clear all overrides
    fetch('/api/config', {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    }).then(() => { showToast('↺ Endpoints reset to defaults', 'success'); poll(); });
  });
}

// ── Maintenance ────────────────────────────────────────────────────────────
function initMaintenance() {
  document.getElementById('maint-db-test')?.addEventListener('click', checkDBStatus);
  document.getElementById('maint-db-migrate')?.addEventListener('click', runDBMigrate);
  document.getElementById('maint-db-purge')?.addEventListener('click', purgeDB);
  document.getElementById('maint-queue-rewarm')?.addEventListener('click', () => rewarmQueue('maint-queue-rewarm'));
  document.getElementById('maint-queue-purge')?.addEventListener('click', purgeQueue);
  // Banner re-warm button (in the pipeline page dropped-warning)
  document.getElementById('rewarm-btn')?.addEventListener('click', () => rewarmQueue('rewarm-btn'));
}

function setMaintBadge(id, status) {
  const el = document.getElementById(id);
  if (!el) return;
  el.className = 'maint-badge ' + status;
  el.textContent = status === 'ok' ? 'Connected' : status === 'error' ? 'Error' : '–';
}

function setMaintResult(id, text, type) {
  const el = document.getElementById(id);
  if (!el) return;
  el.className = 'maint-result ' + type;
  el.textContent = text;
}

async function checkDBStatus() {
  const btn = document.getElementById('maint-db-test');
  if (btn) { btn.disabled = true; btn.textContent = 'Checking…'; }
  try {
    const r = await fetch('/api/maintenance/db/status');
    const d = await r.json();
    if (d.connected && d.schemaError) {
      // Connected but the face_queue table is missing/broken — surface it
      // instead of rendering "schema OK" with zeroed stats.
      setMaintBadge('maint-db-badge', 'error');
      const stats = document.getElementById('maint-db-stats');
      if (stats) stats.classList.remove('visible');
      setMaintResult('maint-db-result',
        `Connected (${d.latencyMs}ms) — schema error: ${d.schemaError}. Run Migration to create the face_queue table.`,
        'error');
    } else if (d.connected) {
      setMaintBadge('maint-db-badge', 'ok');
      const stats = document.getElementById('maint-db-stats');
      if (stats) {
        stats.innerHTML = `
          <strong>Latency:</strong> ${d.latencyMs}ms &nbsp;
          <strong>Pending:</strong> ${fmt(d.pending ?? 0)} &nbsp;
          <strong>Queued:</strong> ${fmt(d.queued ?? 0)} &nbsp;
          <strong>Acknowledged:</strong> ${fmt(d.acknowledged ?? 0)}
        `;
        stats.classList.add('visible');
      }
      setMaintResult('maint-db-result', `Connected (${d.latencyMs}ms) — schema OK`, 'success');
    } else {
      setMaintBadge('maint-db-badge', 'error');
      setMaintResult('maint-db-result', 'Error: ' + d.error, 'error');
    }
  } catch (e) {
    setMaintBadge('maint-db-badge', 'error');
    setMaintResult('maint-db-result', 'Request failed: ' + e.message, 'error');
  } finally {
    if (btn) { btn.disabled = false; btn.textContent = 'Test Connection'; }
  }
}

async function runDBMigrate() {
  const btn = document.getElementById('maint-db-migrate');
  if (btn) { btn.disabled = true; btn.textContent = 'Running…'; }
  setMaintResult('maint-db-result', '', '');
  try {
    const r = await fetch('/api/maintenance/db/migrate', { method: 'POST' });
    const d = await r.json();
    if (d.ok) {
      setMaintResult('maint-db-result', '✓ ' + d.message, 'success');
      showToast('✓ Migration complete', 'success');
      // Refresh DB status after migration
      await checkDBStatus();
    } else {
      setMaintResult('maint-db-result', 'Error: ' + d.error, 'error');
      showToast('Migration failed', 'error');
    }
  } catch (e) {
    setMaintResult('maint-db-result', 'Request failed: ' + e.message, 'error');
    showToast('Migration request failed', 'error');
  } finally {
    if (btn) { btn.disabled = false; btn.textContent = 'Run Migration'; }
  }
}

async function rewarmQueue(btnId) {
  const btn = document.getElementById(btnId);
  if (btn) { btn.disabled = true; btn.textContent = 'Warming…'; }

  // Read current pipeline stats to compute stranded row count before triggering
  let strandedCount = null;
  try {
    const pl = await fetchJSON('/api/pipeline');
    const m = pl.mysql, q = pl.queue;
    if (m?.available && q?.available && m.queued != null) {
      strandedCount = Math.max(0, (m.queued ?? 0) - q.depth);
    }
  } catch (_) {}

  try {
    const r = await fetch('/api/controls/publisher', {
      method:  'PUT',
      headers: { 'Content-Type': 'application/json' },
      body:    JSON.stringify({ warm: true }),
    });
    const d = await r.json();
    if (r.ok) {
      const toastMsg = strandedCount !== null
        ? `↺ Re-warm triggered — ~${fmt(strandedCount)} stranded rows will be re-queued`
        : '↺ Re-warm triggered — MySQL queued rows are being re-pushed';
      showToast(toastMsg, 'success');
      setMaintResult('maint-queue-result', '✓ Re-warm triggered. Watch DB Queued count fall as rows are re-pushed.', 'success');
    } else {
      showToast('Re-warm failed: ' + (d.error || r.status), 'error');
    }
  } catch (e) {
    showToast('Re-warm request failed: ' + e.message, 'error');
  } finally {
    if (btn) {
      btn.disabled = false;
      btn.textContent = '↺ Re-warm Queue';
    }
    poll();
  }
}

async function purgeDB() {
  const confirmed = window.confirm(
    'Purge all rows from face_queue?\n\n' +
    'This resets the pending / queued / acknowledged counters to zero — useful for a clean demo start.\n\n' +
    'The queue backend and publisher/subscriber are NOT affected. Publishing resumes immediately.'
  );
  if (!confirmed) return;

  const btn = document.getElementById('maint-db-purge');
  if (btn) { btn.disabled = true; btn.textContent = 'Purging…'; }
  setMaintResult('maint-db-result', '', '');
  try {
    const r = await fetch('/api/maintenance/db/purge', { method: 'POST' });
    const d = await r.json();
    if (d.ok) {
      setMaintResult('maint-db-result', `✓ ${d.message}`, 'success');
      showToast(`✓ DB purged (${d.rows_deleted} rows removed)`, 'success');
      await checkDBStatus();
    } else {
      setMaintResult('maint-db-result', 'Error: ' + d.error, 'error');
      showToast('DB purge failed', 'error');
    }
  } catch (e) {
    setMaintResult('maint-db-result', 'Request failed: ' + e.message, 'error');
    showToast('DB purge request failed', 'error');
  } finally {
    if (btn) { btn.disabled = false; btn.textContent = 'Purge DB'; }
  }
}

async function purgeQueue() {
  const confirmed = window.confirm(
    'Purge the queue? All messages will be removed immediately.\n\n' +
    'MySQL rows remain as "queued" and will be re-pushed when face-publisher restarts.'
  );
  if (!confirmed) return;

  const btn = document.getElementById('maint-queue-purge');
  if (btn) { btn.disabled = true; btn.textContent = 'Purging…'; }
  setMaintResult('maint-queue-result', '', '');
  try {
    const r = await fetch('/api/maintenance/queue/purge', { method: 'POST' });
    const d = await r.json();
    if (d.ok) {
      setMaintResult('maint-queue-result', '✓ ' + d.message, 'success');
      showToast('✓ Queue purged', 'success');
      // Update queue stats
      await checkQueueStatus();
    } else {
      setMaintResult('maint-queue-result', 'Error: ' + d.error, 'error');
      showToast('Purge failed', 'error');
    }
  } catch (e) {
    setMaintResult('maint-queue-result', 'Request failed: ' + e.message, 'error');
    showToast('Purge request failed', 'error');
  } finally {
    if (btn) { btn.disabled = false; btn.textContent = 'Purge Queue'; }
  }
}

async function checkQueueStatus() {
  try {
    const r = await fetch('/api/pipeline');
    if (!r.ok) return;
    const d = await r.json();
    const q = d.queue;
    if (!q?.available) {
      setMaintBadge('maint-queue-badge', 'error');
      return;
    }
    setMaintBadge('maint-queue-badge', 'ok');
    const stats = document.getElementById('maint-queue-stats');
    if (stats) {
      stats.innerHTML = `
        <strong>Depth:</strong> ${fmt(q.depth)} &nbsp;
        <strong>Max:</strong> ${fmt(q.maxDepth)} &nbsp;
        <strong>Pub rate:</strong> ${q.readyRate ? q.readyRate.toFixed(1)+' msg/s' : '–'} &nbsp;
        <strong>Del rate:</strong> ${q.deliverRate ? q.deliverRate.toFixed(1)+' msg/s' : '–'}
      `;
      stats.classList.add('visible');
    }
  } catch (_) { setMaintBadge('maint-queue-badge', 'error'); }
}

// ── Smiley picker ──────────────────────────────────────────────────────────
let currentEmojiCategory = 0;

async function initSmileyPicker() {
  const tabsEl = document.getElementById('emoji-cat-tabs');

  // Try to load Linky custom images — silently skip if none configured
  let linkyFiles = [];
  try {
    const r = await fetch('/api/linkys');
    if (r.ok) linkyFiles = await r.json();
  } catch (_) {}

  // Build combined category list: emoji categories + optional Linkerd category
  const allCategories = [...EMOJI_CATEGORIES];
  if (linkyFiles.length > 0) {
    allCategories.push({ name: 'Linkerd', icon: '🦞', emojis: null, linkyFiles });
  }

  allCategories.forEach((cat, i) => {
    const tab = document.createElement('button');
    tab.className = 'emoji-cat-tab' + (i === 0 ? ' active' : '');
    tab.title = cat.name;
    tab.textContent = cat.icon;
    tab.addEventListener('click', () => switchEmojiCategory(i, allCategories));
    tabsEl.appendChild(tab);
  });

  // Store for switchEmojiCategory to use
  window._allEmojiCategories = allCategories;
  renderEmojiGrid(0, allCategories);

  // Wire up emoji search input
  const searchInput = document.getElementById('emoji-search-input');
  if (searchInput) {
    searchInput.addEventListener('input', () => {
      const q = searchInput.value.trim();
      if (q) {
        tabsEl.style.display = 'none';
        renderEmojiSearchResults(searchEmojis(q));
      } else {
        tabsEl.style.display = '';
        renderEmojiGrid(currentEmojiCategory, allCategories);
      }
    });
  }

  const ci = document.getElementById('custom-emoji-input');
  ci.addEventListener('input', () => {
    const raw = ci.value.trim(); if (!raw) return;
    clearEmojiGridSelection();
    state.selectedSmiley = raw;
    updateControlsPreview();
  });
}

function switchEmojiCategory(i, cats) {
  currentEmojiCategory = i;
  const allCats = cats || window._allEmojiCategories || EMOJI_CATEGORIES;
  document.querySelectorAll('.emoji-cat-tab').forEach((t, j) => t.classList.toggle('active', j === i));
  renderEmojiGrid(i, allCats);
}

function renderEmojiGrid(catIdx, cats) {
  const allCats = cats || window._allEmojiCategories || EMOJI_CATEGORIES;
  const cat = allCats[catIdx];
  const grid = document.getElementById('emoji-picker-grid');
  grid.innerHTML = '';

  if (cat.linkyFiles) {
    // Linkerd image category — render <img> thumbnails
    for (const filename of cat.linkyFiles) {
      const btn = document.createElement('button');
      btn.className = 'emoji-pick-btn linky-pick-btn';
      btn.title = linkyLabel(filename);
      const img = document.createElement('img');
      img.src = '/linkys/' + filename;
      img.alt = btn.title;
      img.style.cssText = 'width:100%;height:100%;object-fit:contain;pointer-events:none';
      btn.appendChild(img);
      btn.addEventListener('click', async () => {
        clearEmojiGridSelection();
        btn.classList.add('selected');
        btn.textContent = '⏳';
        document.getElementById('custom-emoji-input').value = '';
        try {
          const smileyValue = await linkyToSmileyValue(filename);
          state.selectedSmiley = smileyValue;
          // Restore the img after encoding
          btn.textContent = '';
          btn.appendChild(img);
          btn.classList.add('selected');
          updateControlsPreview();
        } catch (e) {
          showToast('Failed to encode image: ' + e.message, 'error');
          btn.textContent = '❌';
        }
      });
      grid.appendChild(btn);
    }
  } else {
    // Regular emoji category
    for (const emoji of cat.emojis) {
      const btn = document.createElement('button');
      btn.className = 'emoji-pick-btn';
      btn.textContent = emoji;
      btn.title = emoji;
      if (emoji === state.selectedSmiley) btn.classList.add('selected');
      btn.addEventListener('click', () => {
        clearEmojiGridSelection();
        btn.classList.add('selected');
        document.getElementById('custom-emoji-input').value = '';
        state.selectedSmiley = emoji;
        updateControlsPreview();
      });
      grid.appendChild(btn);
    }
  }
}

// Friendly display name for a linky file: "Party_Linky.gif" → "Party Linky"
function linkyLabel(filename) {
  return filename.replace(/\.[^.]+$/, '').replace(/_/g, ' ');
}

// Convert a Linky image file to an <img src="data:..."> smiley value.
// PNGs are canvas-resized to 96×96 to keep payload small.
// GIFs are encoded raw (preserves animation).
// The friendly name rides along in alt/title so every consumer — admin pill
// tooltips, the faces GUI — can show which linky this is after the round trip.
async function linkyToSmileyValue(filename) {
  const url = '/linkys/' + filename;
  const isGif = filename.toLowerCase().endsWith('.gif');

  let dataUri;
  if (isGif) {
    // Fetch raw bytes to preserve animation
    const resp = await fetch(url);
    if (!resp.ok) throw new Error('HTTP ' + resp.status);
    const blob = await resp.blob();
    dataUri = await new Promise((res, rej) => {
      const fr = new FileReader();
      fr.onload = () => res(fr.result);
      fr.onerror = rej;
      fr.readAsDataURL(blob);
    });
  } else {
    // Canvas-resize PNG to 96×96 to minimise payload
    dataUri = await new Promise((res, rej) => {
      const img = new Image();
      img.crossOrigin = 'anonymous';
      img.onload = () => {
        const canvas = document.createElement('canvas');
        canvas.width = canvas.height = 96;
        const ctx = canvas.getContext('2d');
        // Fit within 96×96 preserving aspect ratio, centred
        const scale = Math.min(96 / img.naturalWidth, 96 / img.naturalHeight);
        const w = img.naturalWidth  * scale;
        const h = img.naturalHeight * scale;
        ctx.drawImage(img, (96 - w) / 2, (96 - h) / 2, w, h);
        res(canvas.toDataURL('image/png'));
      };
      img.onerror = rej;
      img.src = url;
    });
  }

  // Use explicit pixel dimensions — the faces-gui cell-smiley is an inline <span>
  // and percentage sizing (width:100%;height:100%) resolves to 0×0 on inline elements.
  // 90px fits comfortably inside the 120px cell without overflow.
  const label = esc(linkyLabel(filename));
  return `<img src="${dataUri}" width="90" height="90" alt="${label}" title="${label}" style="vertical-align:middle;border-radius:4px">`;
}

function clearEmojiGridSelection() {
  document.querySelectorAll('.emoji-pick-btn').forEach(b => b.classList.remove('selected'));
}

// Selected pod IPs — empty array means "all pods"
// Separate selection state for each service — empty array = all pods of that service
let selectedSmileyPods    = [];
let selectedColorPods     = [];
let selectedEmojivotoPods = [];
let evApplyTarget         = 'all'; // 'all' | 'center' | 'edge'
let evApplyDirty          = false; // toggle changed but not yet saved

function initApplyControls() {
  document.querySelectorAll('#apply-target-toggle .toggle-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('#apply-target-toggle .toggle-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      state.applyTarget = btn.dataset.target;
    });
  });
  document.getElementById('apply-controls').addEventListener('click', applyChanges);
  document.getElementById('clear-controls').addEventListener('click', clearControls);
}

async function refreshPodSelector() {
  try {
    // Fetch pod lists and current serving state in parallel
    const [smileyPods, colorPods, smileyState, colorState] = await Promise.all([
      fetchJSON('/api/smileypods').catch(() => []),
      fetchJSON('/api/colorpods').catch(()  => []),
      fetchJSON('/api/smileystate').catch(() => []),
      fetchJSON('/api/colorstate').catch(()  => []),
    ]);

    // Build IP → serving state maps
    const smileyStateMap = Object.fromEntries((smileyState || []).map(s => [s.ip, s]));
    const colorStateMap  = Object.fromEntries((colorState  || []).map(s => [s.ip, s]));

    renderServicePods('pod-sel-smiley-section', 'pod-sel-smiley-row', smileyPods, 'smiley', smileyStateMap);
    renderServicePods('pod-sel-color-section',  'pod-sel-color-row',  colorPods,  'color',  colorStateMap);
  } catch (_) {}
}

// Normalise a pod entry — accepts both plain IP string and PodTopology object
function normPod(p) {
  if (typeof p === 'string') return { ip: p, name: p, zone: null, region: null, node: null, workloadType: null };
  return {
    ip:           p.ip,
    name:         p.name || p.ip,
    zone:         p.zone         || null,
    region:       p.region       || null,
    node:         p.node         || null,
    workloadType: p.workloadType || null,
  };
}

// Multi-line tooltip text for a pod (pre-line CSS renders \n as real line breaks)
function podTooltipText(pod) {
  return [
    pod.name && pod.name !== pod.ip ? `Pod:    ${pod.name}` : null,
    `IP:     ${pod.ip}`,
    pod.zone         ? `Zone:   ${pod.zone}`           : null,
    pod.region       ? `Region: ${pod.region}`         : null,
    pod.node         ? `Node:   ${pod.node}`           : null,
    pod.workloadType === 'externalworkload' ? `Type:   External Workload` : null,
  ].filter(Boolean).join('\n');
}

// Zone-grouped pod selector — groups pods by topology.kubernetes.io/zone.
// Pods without a zone are shown under "On-Premise".
// Zone-level select/deselect available alongside individual pod selection.
// Decode an HTML entity smiley (e.g. "&#x1F603;") to its actual glyph
function decodeEntity(html) {
  if (!html) return '';
  const tmp = document.createElement('span');
  tmp.innerHTML = html;
  return tmp.textContent || tmp.innerText || '';
}

// Extract the data: URI from an "<img …>" smiley value (custom/linky images) so
// display sites can render a size-constrained thumbnail instead of '?'.
// Returns null when the value isn't an image tag with an inline data URI.
// (template.content is inert — nothing loads or executes during parsing.)
function smileyImgSrc(value) {
  if (!value || !value.startsWith('<')) return null;
  const t = document.createElement('template');
  t.innerHTML = value;
  const img = t.content.querySelector('img');
  const src = img && img.getAttribute('src');
  return src && src.startsWith('data:image/') ? src : null;
}

// Friendly name embedded in an "<img …>" smiley value's alt attribute
// (written by linkyToSmileyValue). Null when absent or not an image value.
function smileyImgLabel(value) {
  if (!value || !value.startsWith('<')) return null;
  const t = document.createElement('template');
  t.innerHTML = value;
  const img = t.content.querySelector('img');
  return (img && img.getAttribute('alt')) || null;
}

// Build a small serving-status row to embed in each pod card
function servingStatusEl(serving, service) {
  const el = document.createElement('div');
  el.className = 'pod-serving-row';
  if (!serving) {
    el.innerHTML = `<span class="pod-serving-loading">…</span>`;
    return el;
  }
  if ((service === 'smiley' || service === 'emojivoto') && serving.smiley) {
    // Show center + edge side by side when they differ (label which is which on hover).
    const mkEmoji = (entity, label) => {
      const s = document.createElement('span');
      s.className = 'pod-serving-emoji';
      const imgSrc = smileyImgSrc(entity);
      if (imgSrc) {
        // Custom image value (e.g. a linky) — render a mini thumbnail
        const img = document.createElement('img');
        img.src = imgSrc;
        img.alt = '';
        img.className = 'pod-serving-img';
        s.appendChild(img);
        // Tooltip carries the linky's name (from the value's alt attribute)
        const name = smileyImgLabel(entity);
        if (name) label = label ? `${name} (${label})` : name;
      } else {
        const glyph = entity.startsWith('<') ? '' : decodeEntity(entity);
        s.textContent = glyph || '?';
      }
      if (label) s.setAttribute('data-tooltip', label);
      return s;
    };
    if (serving.smileyEdge) {
      el.appendChild(mkEmoji(serving.smiley, 'Center'));
      el.appendChild(mkEmoji(serving.smileyEdge, 'Edge'));
    } else {
      el.appendChild(mkEmoji(serving.smiley, ''));
    }
  } else if (service === 'color' && serving.color) {
    const mkDot = (hex, label) => {
      const dot = document.createElement('span');
      dot.className = 'pod-serving-color';
      dot.style.background = hex;
      dot.setAttribute('data-tooltip', label || 'Color');
      return dot;
    };
    if (serving.colorEdge) {
      el.appendChild(mkDot(serving.color, 'Center'));
      el.appendChild(mkDot(serving.colorEdge, 'Edge'));
    } else {
      el.appendChild(mkDot(serving.color, ''));
    }
  } else if (serving.error) {
    el.innerHTML = `<span class="pod-serving-loading" style="color:var(--red)">err</span>`;
  }
  return el;
}

function renderServicePods(sectionId, rowId, pods, service, stateMap = {}) {
  const section = document.getElementById(sectionId);
  const row     = document.getElementById(rowId);
  if (!section || !row) return;

  section.style.display = 'flex';

  const getSelected = () => {
    if (service === 'smiley')    return selectedSmileyPods;
    if (service === 'emojivoto') return selectedEmojivotoPods;
    return selectedColorPods;
  };
  const setSelected = arr => {
    if (service === 'smiley')         selectedSmileyPods    = arr;
    else if (service === 'emojivoto') selectedEmojivotoPods = arr;
    else                              selectedColorPods     = arr;
  };

  row.innerHTML = '';

  if (pods.length === 0) {
    row.innerHTML = `<span class="pod-sel-empty">No pods discovered</span>`;
    setSelected([]);
    return;
  }

  // Group by zone — three-way split matching the infrastructure view:
  //   named zone → AZ card (📍)
  //   ExternalWorkload (workloadType=externalworkload) → EW card (🔗)
  //   everything else → On-Premise card (🏢)
  const groups = new Map();
  pods.forEach(raw => {
    const pod = normPod(raw);
    let key, label, icon, isOnPrem = false;
    if (pod.zone) {
      key = pod.zone; label = pod.zone; icon = '📍';
    } else if (pod.workloadType === 'externalworkload') {
      key = '__ew__'; label = 'External Workload'; icon = '🔗'; isOnPrem = true;
    } else {
      key = '__onprem__'; label = 'On-Premise'; icon = '🏢'; isOnPrem = true;
    }
    if (!groups.has(key)) {
      groups.set(key, { label, icon, isOnPrem, pods: [] });
    }
    groups.get(key).pods.push(pod);
  });

  // Named zones alphabetically; External Workload before On-Premise; both last
  const sortedKeys = [...groups.keys()].sort((a, b) => {
    const order = v => v === '__onprem__' ? 2 : v === '__ew__' ? 1 : 0;
    if (order(a) !== order(b)) return order(a) - order(b);
    return a.localeCompare(b);
  });

  // "All pods" button
  const allBtn = document.createElement('button');
  allBtn.className = 'pod-zone-all-btn';
  allBtn.textContent = 'All pods';
  allBtn.addEventListener('click', () => { setSelected([]); rebuild(); });
  row.appendChild(allBtn);

  // Zone card grid
  const grid = document.createElement('div');
  grid.className = 'pod-zone-grid';
  row.appendChild(grid);

  function rebuild() {
    const sel = getSelected();
    allBtn.className = 'pod-zone-all-btn' + (sel.length === 0 ? ' active' : '');
    grid.innerHTML = '';

    sortedKeys.forEach(key => {
      const g       = groups.get(key);
      const sel2    = getSelected();
      const zoneIPs = g.pods.map(p => p.ip);
      const allIn   = zoneIPs.length > 0 && zoneIPs.every(ip => sel2.includes(ip));
      const anyIn   = zoneIPs.some(ip => sel2.includes(ip));

      const card = document.createElement('div');
      card.className = ['pod-zone-card', anyIn ? 'has-selected' : '', allIn ? 'all-selected' : '', g.isOnPrem ? 'is-onprem' : ''].filter(Boolean).join(' ');

      // Header
      const hdr = document.createElement('div');
      hdr.className = 'pod-zone-header';
      const iconEl = document.createElement('span');
      iconEl.className = 'pod-zone-icon'; iconEl.textContent = g.icon;
      const nameEl = document.createElement('span');
      nameEl.className = 'pod-zone-name'; nameEl.textContent = g.label; nameEl.title = g.label;
      const zoneBtn = document.createElement('button');
      zoneBtn.className = 'pod-zone-sel-btn';
      zoneBtn.textContent = allIn ? 'Deselect' : 'Zone';
      zoneBtn.addEventListener('click', e => {
        e.stopPropagation();
        const curr = getSelected();
        setSelected(allIn ? curr.filter(ip => !zoneIPs.includes(ip)) : [...new Set([...curr, ...zoneIPs])]);
        rebuild();
      });
      hdr.append(iconEl, nameEl, zoneBtn);

      // Pod pills
      const podsDiv = document.createElement('div');
      podsDiv.className = 'pod-zone-pods';
      g.pods.forEach(pod => {
        const btn   = document.createElement('button');
        const label = pod.name && pod.name !== pod.ip ? shortPodName(pod.name) : pod.ip.split('.').slice(-2).join('.');
        const tip   = podTooltipText(pod);
        btn.className = 'pod-pill' + (getSelected().includes(pod.ip) ? ' active' : '');
        btn.textContent = label;
        btn.dataset.tooltip = tip; // data-tooltip only — no title attr, avoids double tooltip
        btn.dataset.ip = pod.ip;
        btn.addEventListener('click', () => {
          const curr = getSelected();
          setSelected(curr.includes(pod.ip) ? curr.filter(ip => ip !== pod.ip) : [...curr, pod.ip]);
          rebuild();
        });

        // Inline serving status (emoji or colour swatch) to the right of the pill
        const podRow = document.createElement('div');
        podRow.style.cssText = 'display:flex;align-items:center;gap:4px;';
        podRow.appendChild(btn);
        podRow.appendChild(servingStatusEl(stateMap[pod.ip], service));
        podsDiv.appendChild(podRow);
      });

      card.append(hdr, podsDiv);
      grid.appendChild(card);
    });
  }

  rebuild();
}

// ── Color picker ───────────────────────────────────────────────────────────
function makeSwatch(name, hex, isCustom) {
  const btn = document.createElement('div');
  btn.className = 'color-swatch-btn' + (isCustom ? ' color-swatch-custom' : '');
  btn.dataset.name = isCustom ? '' : name;
  btn.title = isCustom ? 'Custom' : name; // native tooltip on hover — no visible label

  const block = document.createElement('div');
  block.className = 'color-block';
  if (!isCustom) block.style.background = hex;

  btn.appendChild(block);
  return { btn, block };
}

function initColorPicker() {
  const wrap = document.getElementById('color-swatches');

  for (const c of COLOR_PALETTE) {
    const { btn, block } = makeSwatch(c.name, c.hex, false);
    btn.addEventListener('click', () => selectColor(c.name, c.hex, false));
    wrap.appendChild(btn);
  }

  // Custom swatch: rainbow gradient block + native color input
  const { btn: customWrap, block: customBlock } = makeSwatch('Custom', '', true);
  customBlock.innerHTML = `<input type="color" class="color-native-input" value="#53d8fb">`;
  const nativeInput = customBlock.querySelector('input[type="color"]');
  nativeInput.addEventListener('input', e => {
    const hex = e.target.value;
    document.getElementById('color-hex-input').value = hex.slice(1).toUpperCase();
    selectColor(hex, hex, true);
  });
  wrap.appendChild(customWrap);

  const hexInput   = document.getElementById('color-hex-input');
  const hexPreview = document.getElementById('color-hex-preview');
  hexInput.addEventListener('input', () => {
    const raw = hexInput.value.replace(/[^0-9a-fA-F]/g, '').slice(0, 6);
    hexInput.value = raw.toUpperCase();
    if (raw.length === 6) {
      const hex = '#' + raw;
      hexPreview.style.background = hex;
      nativeInput.value = hex;
      selectColor(hex, hex, true);
    } else {
      hexPreview.style.background = 'transparent';
    }
  });
}

function selectColor(name, hex, isCustom = false) {
  state.selectedColor    = name;
  state.selectedColorHex = hex;
  document.querySelectorAll('.color-swatch-btn:not(.color-swatch-custom)').forEach(b => {
    b.classList.toggle('selected', !isCustom && b.dataset.name === name);
  });
  const customSwatch = document.querySelector('.color-swatch-custom');
  if (customSwatch) {
    customSwatch.classList.toggle('selected', isCustom);
    // Update the custom block background to show the chosen color
    if (isCustom) {
      const block = customSwatch.querySelector('.color-block');
      if (block) block.style.background = hex;
    }
  }
  updateControlsPreview();
}

function updateControlsPreview() {
  const cell  = document.getElementById('face-preview-cell');
  const emoji = document.getElementById('preview-emoji');
  const hint  = document.getElementById('face-preview-hint');
  const applyBtn = document.getElementById('apply-controls');

  const hasEmoji = !!state.selectedSmiley;
  const hasColor = !!state.selectedColorHex;

  // Background
  const bg = hasColor ? state.selectedColorHex : 'var(--card-bg)';
  cell.style.backgroundColor = bg;
  // Text readable against background
  if (hasColor) {
    cell.style.color = isLightColor(state.selectedColorHex) ? '#222' : '#f0f0f0';
  } else {
    cell.style.color = '';
  }

  // Emoji — use innerHTML for HTML entities and <img> tags, textContent otherwise
  if (hasEmoji) {
    const s = state.selectedSmiley;
    if (s.startsWith('<') || s.startsWith('&#')) {
      emoji.innerHTML = s;
    } else {
      emoji.textContent = s;
    }
  } else {
    emoji.textContent = hasColor ? '' : '+';
  }

  // Hint text — custom image values are giant <img> markup; show their name instead
  const parts = [];
  if (hasEmoji) {
    parts.push(state.selectedSmiley.startsWith('<')
      ? (smileyImgLabel(state.selectedSmiley) || 'custom image')
      : state.selectedSmiley);
  }
  if (hasColor) parts.push(state.selectedColorHex);
  hint.textContent = parts.length ? parts.join(' on ') : 'Select an emoji or color';

  applyBtn.disabled = !hasEmoji && !hasColor;
}

function clearControls() {
  state.selectedSmiley    = null;
  state.selectedColor     = null;
  state.selectedColorHex  = null;
  // Deselect any highlighted emoji/color swatch buttons
  document.querySelectorAll('.emoji-pick-btn.selected').forEach(b => b.classList.remove('selected'));
  document.querySelectorAll('.color-swatch-btn.selected').forEach(b => b.classList.remove('selected'));
  updateControlsPreview();
}

async function applyChanges() {
  const hasEmoji = !!state.selectedSmiley;
  const hasColor = !!state.selectedColorHex;
  if (!hasEmoji && !hasColor) return;

  const btn = document.getElementById('apply-controls');
  btn.disabled = true; btn.textContent = 'Applying…';

  try {
    const tasks = [];
    if (hasEmoji) tasks.push(fetch('/api/smiley', {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        which: state.applyTarget,
        smiley: state.selectedSmiley,
        ...(selectedSmileyPods.length > 0 ? { pods: selectedSmileyPods } : {}),
      }),
    }));
    if (hasColor) tasks.push(fetch('/api/color', {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        which: state.applyTarget,
        color: state.selectedColorHex,  // always send hex — server accepts #rrggbb via passthrough
        ...(selectedColorPods.length > 0 ? { pods: selectedColorPods } : {}),
      }),
    }));
    const labels  = [hasEmoji && 'emoji', hasColor && 'color'].filter(Boolean);
    const results = await Promise.all(tasks);

    // Collect per-request outcomes: hard failures carry the server's error text
    // (e.g. "smiley rejected: unknown smiley" from an older workload build);
    // OK responses may still report per-pod partial failure via succeeded/pods.
    const errors = [];
    let partial = null;
    for (let i = 0; i < results.length; i++) {
      const r = results[i];
      if (!r.ok) {
        const text = (await r.text().catch(() => '')).trim();
        errors.push(`${labels[i]}: ${text || r.statusText}`);
        continue;
      }
      const d = await r.json().catch(() => null);
      if (d && d.pods > 0 && d.succeeded < d.pods && !partial) {
        partial = { label: labels[i], succeeded: d.succeeded, pods: d.pods };
      }
    }

    if (errors.length > 0) {
      showToast('Apply failed — ' + errors.join('; '), 'error');
    } else if (partial) {
      const rejected = partial.pods - partial.succeeded;
      showToast(`⚠ ${partial.label} applied to ${partial.succeeded}/${partial.pods} pods — ` +
                `${rejected} workload${rejected !== 1 ? 's' : ''} rejected it (older faces build?)`, 'warn');
      refreshPodSelector();
    } else {
      const podNote = (selectedSmileyPods.length > 0 || selectedColorPods.length > 0)
        ? ` (targeted pods)`
        : '';
      showToast(`✓ Applied ${labels.join(' + ')} → ${state.applyTarget}${podNote}`, 'success');
      // Immediately re-fetch serving state so pod pills reflect the new value
      refreshPodSelector();
    }
  } catch (e) {
    showToast('Apply failed: ' + e.message, 'error');
  } finally {
    btn.disabled = false; btn.textContent = 'Apply';
  }
}

// ── Live View ──────────────────────────────────────────────────────────────
// ── Live View — shared state ───────────────────────────────────────────────
let liveRunning  = false, liveTimer = null;
let liveRows = 8, liveCols = 8, liveInterval = 2000;

// Rolling rate tracking (5-second window)
let liveConsumedTotal = 0;
const RATE_WINDOW_MS  = 5000;
const liveRateWindow  = []; // [{ts, count}]

// Rate (per second) over the trailing windowMs of {ts, count} samples.
// Counts after the window-start sample over the elapsed time since it.
// Mutates win in place to drop expired samples.
function windowedRate(win, now, windowMs) {
  while (win.length > 0 && win[0].ts < now - windowMs) win.shift();
  if (win.length < 2) return 0;
  const total = win.slice(1).reduce((s, e) => s + e.count, 0);
  const elapsed = (win.at(-1).ts - win[0].ts) / 1000;
  return elapsed > 0 ? total / elapsed : 0;
}

function trackLiveConsumed(count) {
  // Zero-count ticks are recorded too — otherwise the rate display would
  // freeze at its last non-zero value when consumption stops (empty queue).
  const now = Date.now();
  liveConsumedTotal += count;
  liveRateWindow.push({ ts: now, count });
  const rate = windowedRate(liveRateWindow, now, RATE_WINDOW_MS);

  const rateEl = document.getElementById('live-rate-val');
  const consEl = document.getElementById('live-consumed-val');
  if (rateEl) rateEl.textContent = rate > 0 ? rate.toFixed(1) : '–';
  if (consEl) consEl.textContent = fmt(liveConsumedTotal);
}
let currentLvMode = 'grid'; // 'grid' | 'drain'

function initGridPicker() {
  const PICKER_MAX  = 12;
  const PICKER_CELL = 14; // px — slightly larger cells since picker is smaller now
  const popup  = document.getElementById('grid-picker-popup');
  const gpGrid = document.getElementById('grid-picker-grid');
  const info   = document.getElementById('grid-picker-info');
  const btn    = document.getElementById('grid-picker-btn');

  // Build the PICKER_MAX × PICKER_MAX cell grid once
  gpGrid.style.gridTemplateColumns = `repeat(${PICKER_MAX}, ${PICKER_CELL}px)`;
  for (let r = 1; r <= PICKER_MAX; r++) {
    for (let c = 1; c <= PICKER_MAX; c++) {
      const cell = document.createElement('div');
      cell.className = 'gp-cell';
      cell.dataset.row = r;
      cell.dataset.col = c;
      gpGrid.appendChild(cell);
    }
  }

  function highlight(rows, cols) {
    gpGrid.querySelectorAll('.gp-cell').forEach(cell => {
      cell.classList.toggle('selected',
        Number(cell.dataset.row) <= rows && Number(cell.dataset.col) <= cols);
    });
    info.textContent = `${cols} × ${rows}`;
  }

  // Resolve the hovered cell from the cursor position. The 1px gaps between
  // cells are dead zones for e.target.closest() — landing on one would freeze
  // the highlight mid-sweep — so compute row/col geometrically instead.
  const STRIDE = PICKER_CELL + 1; // cell + 1px gap
  function cellAt(clientX, clientY) {
    const r = gpGrid.getBoundingClientRect();
    const col = Math.min(PICKER_MAX, Math.max(1, Math.ceil((clientX - r.left) / STRIDE)));
    const row = Math.min(PICKER_MAX, Math.max(1, Math.ceil((clientY - r.top) / STRIDE)));
    return { row, col };
  }

  gpGrid.addEventListener('mousemove', e => {
    const { row, col } = cellAt(e.clientX, e.clientY);
    highlight(row, col);
  });

  gpGrid.addEventListener('mouseleave', () => highlight(liveRows, liveCols));

  gpGrid.addEventListener('click', e => {
    // Geometry-based like mousemove, so clicking a gap still selects a size.
    const { row, col } = cellAt(e.clientX, e.clientY);
    const was = liveRunning; if (was) stopLive();
    buildLiveGrid(row, col);
    if (was) startLive();
    popup.hidden = true;
  });

  btn.addEventListener('click', e => {
    e.stopPropagation();
    popup.hidden = !popup.hidden;
    if (!popup.hidden) highlight(liveRows, liveCols);
  });

  document.addEventListener('click', () => { popup.hidden = true; });
  popup.addEventListener('click', e => e.stopPropagation());
}

function initLiveView() {
  buildLiveGrid(liveRows, liveCols);
  initGridPicker();

  // ResizeObserver fires whenever the wrap container gets real dimensions,
  // e.g. on first navigation to Live View or after any layout change.
  // This is more reliable than requestAnimationFrame for post-layout sizing.
  const wrap = document.querySelector('.live-grid-wrap');
  if (wrap && window.ResizeObserver) {
    new ResizeObserver(() => sizeLiveGrid()).observe(wrap);
  }

  // Mode selector
  document.querySelectorAll('.lv-mode-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const mode = btn.dataset.lvMode;
      if (mode === currentLvMode) return;
      // Stop whichever is running before switching
      if (currentLvMode === 'grid')  stopLive();
      if (currentLvMode === 'drain') stopDrain();
      currentLvMode = mode;
      document.querySelectorAll('.lv-mode-btn').forEach(b => b.classList.toggle('active', b.dataset.lvMode === mode));
      document.getElementById('lv-grid-view').classList.toggle('lv-section-active', mode === 'grid');
      document.getElementById('lv-drain-view').classList.toggle('lv-section-active', mode === 'drain');
    });
  });

  // Grid controls
  document.getElementById('live-toggle').addEventListener('click', () => liveRunning ? stopLive() : startLive());
  document.getElementById('live-interval').addEventListener('change', e => {
    liveInterval = parseInt(e.target.value);
    if (liveRunning) { stopLive(); startLive(); }
  });

  // Legend toggle
  const legendBtn   = document.getElementById('live-legend-btn');
  const legendPopup = document.getElementById('live-legend-popup');
  legendBtn.addEventListener('click', e => {
    e.stopPropagation();
    legendPopup.hidden = !legendPopup.hidden;
    legendBtn.classList.toggle('open', !legendPopup.hidden);
  });
  document.addEventListener('click', () => { legendPopup.hidden = true; legendBtn.classList.remove('open'); });
  legendPopup.addEventListener('click', e => e.stopPropagation());

  // Drain controls
  initDrainMode();
  updateLiveViewModeUI();
}

// Default cell size (px) — matches the faces-gui app's cell size.
// Cells shrink below this only when needed to fit all rows×cols in the viewport.
const LIVE_DEFAULT_CELL_PX = 100;
const LIVE_MIN_CELL_PX     = 20;
const LIVE_CELL_GAP        = 4;

// Compute and apply the largest square cell size that fits the live-grid-wrap
// without scrolling, capped at LIVE_DEFAULT_CELL_PX.
// Guards against running before layout is complete (clientHeight === 0).
function sizeLiveGrid() {
  const wrap = document.querySelector('.live-grid-wrap');
  const grid = document.getElementById('live-grid');
  if (!wrap || !grid || !liveRows || !liveCols) return;

  const availW = wrap.clientWidth;
  const availH = wrap.clientHeight;
  if (availW <= 0 || availH <= 0) return; // layout not ready — ResizeObserver will retry

  const gapW = LIVE_CELL_GAP * (liveCols - 1);
  const gapH = LIVE_CELL_GAP * (liveRows  - 1);

  const maxByW = Math.floor((availW - gapW) / liveCols);
  const maxByH = Math.floor((availH - gapH) / liveRows);

  const cellPx  = Math.max(LIVE_MIN_CELL_PX, Math.min(LIVE_DEFAULT_CELL_PX, maxByW, maxByH));
  const fontPx  = Math.max(8, Math.round(cellPx * 0.62)); // ~62% of cell = emoji fits with padding

  grid.style.gridTemplateColumns = `repeat(${liveCols}, ${cellPx}px)`;
  grid.style.gridTemplateRows    = `repeat(${liveRows},  ${cellPx}px)`;
  grid.style.gap                 = `${LIVE_CELL_GAP}px`;
  grid.style.setProperty('--live-font-px', fontPx + 'px');
  grid.classList.add('live-grid--sized'); // reveal after first successful sizing
}

function buildLiveGrid(rows, cols) {
  liveRows = rows; liveCols = cols;
  // Keep the picker button label in sync
  const lbl = document.getElementById('grid-size-label');
  if (lbl) lbl.textContent = `${cols} × ${rows}`;
  const grid = document.getElementById('live-grid');
  grid.innerHTML = '';
  grid.classList.remove('live-grid--sized'); // hide until re-sized
  grid.style.setProperty('--live-cols', cols);
  grid.style.setProperty('--live-rows', rows);
  for (let r = 0; r < rows; r++)
    for (let c = 0; c < cols; c++) {
      const cell = document.createElement('div');
      cell.className = 'live-cell loading';
      cell.id = `lc-${r}-${c}`;
      cell.textContent = '·';
      grid.appendChild(cell);
    }
  sizeLiveGrid(); // may no-op if wrap has no height yet; ResizeObserver retries
  setLiveNote();
}

function setLiveNote() {
  const note = document.getElementById('live-note');
  if (!note) return;
  note.textContent = state.mode === 'pubsub'
    ? `Each cell is a live consumer — in pub/sub mode this reads from the queue. Outer cells → edge endpoint, inner cells → center endpoint. ${liveRows * liveCols} polls per tick.`
    : `Outer cells → edge endpoint, inner cells → center endpoint. ${liveRows * liveCols} concurrent polls per tick.`;
  // "consumed" only makes sense against a queue; classic mode just serves requests
  const lbl = document.getElementById('live-consumed-lbl');
  if (lbl) lbl.textContent = state.mode === 'pubsub' ? 'consumed' : 'served';
}

function startLive() {
  // Reset counters on each new session
  liveConsumedTotal = 0;
  liveRateWindow.length = 0;
  liveRunning = true;
  updateLiveBtn();
  tickLiveView();
  liveTimer = setInterval(tickLiveView, liveInterval);
}
function stopLive() {
  liveRunning = false; clearInterval(liveTimer); liveTimer = null; updateLiveBtn();
}
function updateLiveBtn() {
  const btn    = document.getElementById('live-toggle');
  const st     = document.getElementById('live-status');
  const rateSt = document.getElementById('live-rate-stat');
  const consSt = document.getElementById('live-consumed-stat');
  if (!btn) return;
  if (liveRunning) {
    btn.textContent = '⏸ Pause'; btn.className = 'live-toggle-btn pause';
    st.textContent = `● ${liveInterval/1000}s interval`; st.className = 'live-status running';
    if (rateSt) rateSt.style.display = '';
    if (consSt) consSt.style.display = '';
  } else {
    btn.textContent = '▶ Start'; btn.className = 'live-toggle-btn start';
    st.textContent = 'stopped'; st.className = 'live-status';
    if (rateSt) rateSt.style.display = 'none';
    if (consSt) consSt.style.display = 'none';
  }
}

async function tickLiveView() {
  const cells = [];
  for (let r = 0; r < liveRows; r++)
    for (let c = 0; c < liveCols; c++) {
      const isEdge = (r === 0 || r === liveRows - 1 || c === 0 || c === liveCols - 1);
      cells.push({ id: `lc-${r}-${c}`, row: r, col: c, endpoint: isEdge ? 'edge' : 'center' });
    }

  const results = await Promise.allSettled(
    cells.map(cell =>
      fetch(`/face/${cell.endpoint}?row=${cell.row}&col=${cell.col}`).then(r => r.json())
        .then(d => ({ ...d, cellId: cell.id }))
        .catch(err => ({ cellId: cell.id, smiley: '&#x1F92C;', color: '#BBBBBB', errors: [err.message] }))
    )
  );

  let tickConsumed = 0;
  for (const result of results) {
    if (result.status === 'rejected') continue;
    const d = result.value;
    const el = document.getElementById(d.cellId); if (!el) continue;
    el.classList.remove('loading', 'empty', 'error');
    const imgSrc = smileyImgSrc(d.smiley || '');
    if (imgSrc) {
      el.textContent = '';
      const img = document.createElement('img');
      img.src = imgSrc;
      img.alt = '';
      img.className = 'live-cell-img';
      el.appendChild(img);
    } else {
      const tmp = document.createElement('span'); tmp.innerHTML = d.smiley || '&#x1F92C;';
      el.textContent = tmp.textContent || '?';
    }
    const color = d.color || '#BBBBBB';
    el.style.backgroundColor = color;
    el.style.color = isLightColor(color) ? '#222' : '#f0f0f0';
    el.style.borderColor = 'transparent';
    if (d.status === 503) el.classList.add('empty');
    else if (d.errors?.length) el.classList.add('error');
    // Only a 200 delivered a message — a 503 means "queue empty", nothing consumed
    if (d.status === 200) tickConsumed++;
  }
  trackLiveConsumed(tickConsumed);
}

// ── Drain mode ─────────────────────────────────────────────────────────────
let drainRunning = false;
let drainWorkers = 4;
let drainSuccess  = 0; // 200s — messages actually consumed from the queue
let drainEmpty    = 0; // 503s — queue was empty, nothing consumed
let drainFailed   = 0;
let drainStartMs = null;
const drainWindow = []; // [{ts, count}] for the rolling msg/s rate

function initDrainMode() {
  const slider = document.getElementById('drain-workers-slider');
  const valEl  = document.getElementById('drain-workers-val');

  slider.addEventListener('input', () => {
    drainWorkers = Number(slider.value);
    const fill = Math.round(((drainWorkers - 1) / 31) * 100);
    slider.style.setProperty('--fill', fill + '%');
    valEl.textContent = drainWorkers;
  });

  document.getElementById('drain-toggle').addEventListener('click', () => {
    drainRunning ? stopDrain() : startDrain();
  });
}

const DRAIN_MAX_WORKERS = 32;
let drainGen = 0; // generation token — invalidates workers from a previous run

function startDrain() {
  if (drainRunning) return;
  drainRunning = true;
  drainSuccess  = 0;
  drainEmpty    = 0;
  drainFailed   = 0;
  drainWindow.length = 0;
  drainStartMs  = Date.now();
  updateDrainUI();
  // Independent continuous workers — each fires its next request the moment
  // the previous one completes. No batch barrier, so one slow (chaos-delayed)
  // response no longer stalls the rest. Workers above the slider value park
  // until the slider is raised, which keeps live adjustment working mid-run.
  const gen = ++drainGen;
  for (let i = 0; i < DRAIN_MAX_WORKERS; i++) drainWorkerLoop(i, gen);
}

function stopDrain() {
  drainRunning = false;
  updateDrainUI();
}

async function drainWorkerLoop(idx, gen) {
  while (drainRunning && gen === drainGen) {
    if (idx >= drainWorkers) {
      await new Promise(r => setTimeout(r, 200)); // parked — slider is below our index
      continue;
    }
    const row = Math.floor(idx / 8), col = idx % 8;
    let got200 = false, got503 = false;
    try {
      const r = await fetch(`/face/center?row=${row}&col=${col}`);
      if (r.status === 200)      { drainSuccess++; got200 = true; }
      else if (r.status === 503) { drainEmpty++; got503 = true; } // queue empty / unavailable
      else                       { drainFailed++; }
    } catch (_) { drainFailed++; }
    drainWindow.push({ ts: Date.now(), count: got200 ? 1 : 0 });
    scheduleDrainUI();
    // Idle pacing per worker: on an empty queue poll gently instead of
    // hammering 503s; delivery resumes at full speed on the next message.
    if (drainRunning && got503) await new Promise(r => setTimeout(r, 500));
  }
}

// Repaint at most 4×/s — with continuous workers the per-request updates
// would otherwise thrash the DOM at hundreds of updates per second.
let drainUiLast = 0;
function scheduleDrainUI() {
  const now = Date.now();
  if (now - drainUiLast >= 250) {
    drainUiLast = now;
    updateDrainUI();
  }
}

function updateDrainUI() {
  const btn    = document.getElementById('drain-toggle');
  const badge  = document.getElementById('drain-badge');
  const active = document.getElementById('drain-workers-active');
  if (!btn) return;

  const isPubsub = state.mode === 'pubsub';
  if (drainRunning) {
    btn.textContent = isPubsub ? '■ Stop Drain' : '■ Stop Load'; btn.className = 'drain-toggle-btn stop';
    badge.textContent = isPubsub ? 'Draining' : 'Running'; badge.className = 'drain-badge running';
    active.textContent = `${drainWorkers} workers`;
  } else {
    btn.textContent = isPubsub ? '⚡ Start Drain' : '⚡ Start Load'; btn.className = 'drain-toggle-btn start';
    badge.textContent = 'Stopped'; badge.className = 'drain-badge';
    active.textContent = '';
  }

  // Stats — "consumed" is real deliveries (200s) only; 503s are just empty polls
  document.getElementById('drain-consumed').textContent = fmt(drainSuccess);
  document.getElementById('drain-empty').textContent    = fmt(drainEmpty);
  document.getElementById('drain-failed').textContent   = fmt(drainFailed);
  document.getElementById('drain-attempts').textContent = fmt(drainSuccess + drainEmpty + drainFailed);
  const elapsed = drainStartMs ? Math.floor((Date.now() - drainStartMs) / 1000) : 0;
  document.getElementById('drain-elapsed').textContent = drainStartMs
    ? `${Math.floor(elapsed / 60)}:${String(elapsed % 60).padStart(2, '0')}`
    : '–';
  // Rolling window so the rate reflects NOW, not a since-start average —
  // it decays to zero once the queue is dry instead of tapering slowly.
  const rate = windowedRate(drainWindow, Date.now(), RATE_WINDOW_MS);
  document.getElementById('drain-rate').textContent = rate > 0 ? rate.toFixed(1) : '–';

  const qDepth = state.lastPipeline?.queue?.depth ?? null;
  document.getElementById('drain-queue').textContent = qDepth !== null ? fmt(qDepth) : '–';
  if (qDepth !== null && rate > 0) {
    const eta = qDepth / rate;
    document.getElementById('drain-eta').textContent = qDepth === 0 ? '✓ Empty'
      : eta > 120 ? Math.round(eta / 60) + ' min'
      : Math.round(eta) + 's';
  } else {
    document.getElementById('drain-eta').textContent = qDepth === 0 ? '✓ Empty' : '–';
  }
}

// Classic mode has no queue, so the drain engine doubles as a plain load
// generator — all queue-flavored wording and stats flip with the mode.
function updateLiveViewModeUI() {
  const isPubsub = state.mode === 'pubsub';
  const set = (id, v) => { const el = document.getElementById(id); if (el) el.textContent = v; };

  const modeBtn = document.querySelector('.lv-mode-btn[data-lv-mode="drain"]');
  if (modeBtn) modeBtn.textContent = isPubsub ? '⚡ Drain Mode' : '🔥 Load Mode';

  set('drain-consumed-lbl', isPubsub ? 'consumed (200)' : 'success (200)');
  set('drain-rate-lbl',     isPubsub ? 'msg / sec' : 'req / sec');
  set('drain-empty-lbl',    isPubsub ? 'empty polls (503)' : 'unavailable (503)');

  // Queue depth and drain ETA only exist against a queue
  const queueCell = document.getElementById('drain-stat-queue');
  const etaCell   = document.getElementById('drain-stat-eta');
  if (queueCell) queueCell.style.display = isPubsub ? '' : 'none';
  if (etaCell)   etaCell.style.display   = isPubsub ? '' : 'none';
  const row1 = document.getElementById('drain-stats');
  if (row1) row1.style.gridTemplateColumns = isPubsub ? 'repeat(4,1fr)' : 'repeat(2,1fr)';

  const note = document.getElementById('drain-note');
  if (note) {
    note.textContent = isPubsub
      ? 'Concurrent workers each pull the next message as soon as their previous request completes. Every 200 acknowledges one message from the queue — use this to drain the queue quickly without a full GUI session.'
      : 'Concurrent workers hammer the face endpoint as fast as it responds — a lightweight load generator for driving high request rates without building a full load test.';
  }

  updateDrainUI(); // refresh badge / button wording
  setLiveNote();
}

function isLightColor(hex) {
  if (!hex || hex.length < 7) return false;
  const r = parseInt(hex.slice(1,3),16), g = parseInt(hex.slice(3,5),16), b = parseInt(hex.slice(5,7),16);
  return (r*299 + g*587 + b*114) / 1000 > 128;
}

// ── Toast ──────────────────────────────────────────────────────────────────
let toastTimer;
function showToast(msg, type = 'success') {
  const el = document.getElementById('toast');
  el.textContent = msg;
  el.className = `toast ${type === 'warn' ? 'success' : type} show`;
  if (type === 'warn') el.style.borderLeftColor = 'var(--yellow)';
  else el.style.borderLeftColor = '';
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { el.className = 'toast'; }, 3500);
}

// ── Helpers ────────────────────────────────────────────────────────────────
function fmt(n) { return n == null ? '–' : Number(n).toLocaleString(); }

// Shorten a Kubernetes pod name to its unique suffix.
// "face-publisher-5d5f9b7f8c-xk2j9" → "xk2j9"
function shortPodName(name) {
  if (!name) return name;
  const parts = name.split('-');
  return parts.length >= 3 ? parts[parts.length - 1] : name;
}


// Convert a publish interval in ms to a human-readable messages-per-second string.
function msToRate(ms) {
  if (ms === 0) return 'Flood';
  const rate = 1000 / ms;
  if (rate >= 10) return `${Math.round(rate)} msg/s`;
  if (rate >= 1)  return `${rate.toFixed(1)} msg/s`;
  return `${rate.toFixed(2)} msg/s`;
}

// Logarithmic slider helpers.
// Slider range 0–100:  0 = flood (0 ms),  1–100 maps log from 1 ms to 1000 ms.
// This gives equal visual space to each decade (1–10 ms, 10–100 ms, 100–1000 ms).
const LOG_MAX_MS = 1000;
function sliderToMs(val) {
  if (val <= 0) return 0;
  return Math.round(Math.exp((val / 100) * Math.log(LOG_MAX_MS)));
}
function msToSlider(ms) {
  if (ms <= 0) return 0;
  return Math.round((Math.log(Math.min(ms, LOG_MAX_MS)) / Math.log(LOG_MAX_MS)) * 100);
}
function esc(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}
// ── JS tooltip (position:fixed — works inside overflow:auto containers) ───────
function initTooltips() {
  const tip = document.createElement('div');
  tip.id = 'js-tooltip';
  document.body.appendChild(tip);

  let hideTimer;

  const mkSwatch = (hex) => {
    const s = document.createElement('span');
    s.style.cssText = `display:inline-block;width:12px;height:12px;background:${hex};` +
      `border-radius:3px;border:1px solid rgba(128,128,128,.35);vertical-align:middle`;
    return s;
  };

  function show(text, x, y, colorHex, colorHexEdge) {
    clearTimeout(hideTimer);
    if (colorHex) {
      // Colour pods: swatch(es) are the status block at the top — the hex value is omitted.
      // The #js-tooltip is white-space:pre-line, so "\n" separates rows.
      tip.innerHTML = '';
      if (colorHexEdge) {
        // Center & edge differ — two labelled swatches
        tip.appendChild(document.createTextNode('Center: '));
        tip.appendChild(mkSwatch(colorHex));
        tip.appendChild(document.createTextNode('\nEdge: '));
        tip.appendChild(mkSwatch(colorHexEdge));
        tip.appendChild(document.createTextNode('\n' + text));
      } else {
        tip.appendChild(mkSwatch(colorHex));
        tip.appendChild(document.createTextNode('\n' + text));
      }
    } else {
      tip.textContent = text;
    }
    tip.classList.add('visible');
    move(x, y);
  }
  function move(x, y) {
    const W = window.innerWidth, H = window.innerHeight;
    const w = tip.offsetWidth  || 230;
    const h = tip.offsetHeight || 60;
    let left = x + 14;
    let top  = y - h - 10;
    if (left + w > W - 8) left = x - w - 14;
    if (top < 8)          top  = y + 18;
    tip.style.left = left + 'px';
    tip.style.top  = top  + 'px';
  }
  function hide() {
    clearTimeout(hideTimer);
    hideTimer = setTimeout(() => tip.classList.remove('visible'), 80);
  }

  document.addEventListener('mouseover', e => {
    const el = e.target.closest('[data-tooltip]');
    if (el?.dataset.tooltip) show(el.dataset.tooltip.replace(/\\n/g, '\n'), e.clientX, e.clientY, el.dataset.podColor || null, el.dataset.podColorEdge || null);
  });
  document.addEventListener('mousemove', e => {
    if (tip.classList.contains('visible')) move(e.clientX, e.clientY);
  });
  document.addEventListener('mouseout', e => {
    if (e.target.closest('[data-tooltip]')) hide();
  });
}

function metricCard(value, label, cls='', tooltip='') {
  const tipIcon = tooltip ? `<span class="tip-icon" data-tooltip="${esc(tooltip)}">i</span>` : '';
  return `<div class="metric-card">${tipIcon}<div class="metric-value ${cls}">${esc(String(value))}</div><div class="metric-label">${esc(label)}</div></div>`;
}
// secondary: optional small line below the label (e.g. "2 pending" on the MySQL node)
// secondaryWarn: if true, secondary text gets the orange .warn class
function pipeNode(icon, name, count, label, unavail=false, paused=false, tooltip='', tipPos='', secondary='', secondaryWarn=false) {
  const cls = ['pipe-node', unavail?'unavailable':'', paused?'paused':''].filter(Boolean).join(' ');
  const tipIcon = tooltip ? `<span class="tip-icon" data-tooltip="${esc(tooltip)}">i</span>` : '';
  const secondaryHtml = secondary
    ? `<div class="node-secondary${secondaryWarn?' warn':''}">${esc(secondary)}</div>`
    : '';
  return `<div class="${cls}">${tipIcon}<span class="node-icon">${icon}</span><div class="node-name">${esc(name)}</div><div class="node-count">${esc(String(count))}</div><div class="node-label">${esc(label)}</div>${secondaryHtml}</div>`;
}
function pipeArrow(label, state='') {
  return `<div class="pipe-arrow ${state}"><div class="arrow-label">${esc(label)}</div><div class="arrow-line"></div></div>`;
}

// ── Infrastructure view ────────────────────────────────────────────────────
// Polls /api/infrastructure on a slower cadence (each call pings every pod).
// Renders zone cards that adapt: multi-zone cloud layout when topology labels
// are present, flat on-premise card otherwise.
const INFRA_POLL_MS = 10000;

async function pollInfra() {
  try {
    const infra = await fetchJSON('/api/infrastructure');
    state.lastInfra = infra;          // cache for Fault Injection scope targeting
    renderInfrastructure(infra);
    // Repaint the FI Status & Targeting topology if that page exists
    renderFITopology(infra, state.lastChaos || {}, state.mode);
    // Refresh dynamic outage buttons (zone/node list may have changed)
    renderFIOutageButtons(state.mode);
    // Refresh FI pod-pill glyphs — smiley/color/face pills render once at card
    // creation, usually BEFORE the first (slow) infra poll lands, so their
    // emoji/swatch would otherwise never appear.
    refreshFIPillGlyphs();
  } catch (e) {
    console.warn('infra poll error', e);
  }
}

// Display order for services within a zone card
const INFRA_SVC_ORDER = {
  pubsub:  ['gui', 'subscriber', 'publisher', 'smiley', 'color'],
  classic: ['gui', 'face', 'smiley', 'color'],
};
const INFRA_SVC_META = {
  publisher:  { icon: '📤', label: 'Publisher'  },
  subscriber: { icon: '📥', label: 'Subscriber' },
  smiley:     { icon: '😀', label: 'Smiley'     },
  color:      { icon: '🎨', label: 'Color'       },
  face:       { icon: '😃', label: 'Face'        },
  gui:        { icon: '🖥️', label: 'GUI'         },
};

function renderInfrastructure(infra) {
  const section = document.getElementById('infra-section');
  const content = document.getElementById('infra-content');
  if (!section || !content) return;

  if (!infra || !infra.zones || infra.zones.length === 0) {
    section.style.display = 'none';
    return;
  }
  section.style.display = '';

  const svcOrder = INFRA_SVC_ORDER[infra.mode] || INFRA_SVC_ORDER.classic;

  // ── Group zones by environment type ───────────────────────────────────────
  const regions   = [...new Set(infra.zones.filter(z => z.zone).map(z => z.region).filter(Boolean))];
  const regionStr = regions.length === 1 ? regions[0] : (regions.length > 1 ? 'Multi-Region' : '');

  const cloudZones  = infra.zones.filter(z => z.zone !== '');
  const ewZones     = infra.zones.filter(z => z.label === 'External Workload');
  const opZones     = infra.zones.filter(z => z.label === 'On-Premise');

  // Service dependency priority — lower = more upstream / user-facing.
  // Used to number nodes so the node with the most important service gets "Node 1".
  //   0: gui       — user entry point
  //   1: face/subscriber — request handlers
  //   2: publisher  — background pipeline worker
  //   3: smiley/color — backend leaf services
  const SVC_NODE_PRIORITY = {
    gui: 0, face: 1, subscriber: 1, publisher: 2, smiley: 3, color: 3,
  };
  function nodeScore(svcMap) {
    // A node's score = min priority across all its services (lower = more upstream)
    return Object.keys(svcMap).reduce((min, svc) => {
      const p = SVC_NODE_PRIORITY[svc] ?? 99;
      return p < min ? p : min;
    }, Infinity);
  }

  // For on-prem and EW groups: split one flat card into one-card-per-node when
  // pods carry node info and multiple distinct nodes exist.
  // Nodes are ordered by dependency position (GUI/Face node → "Node 1"; backend node → "Node 2").
  // Hover on the label shows the real node name.
  function subdivideByNode(zones, baseIcon) {
    if (!zones.length) return zones;

    // Flatten every pod across every service in this group
    const nodeGroups = new Map(); // nodeName → { svc → [pod] }
    for (const zone of zones) {
      for (const [svc, pods] of Object.entries(zone.pods || {})) {
        for (const pod of pods) {
          const n = pod.node || '__unknown__';
          if (!nodeGroups.has(n)) nodeGroups.set(n, {});
          const g = nodeGroups.get(n);
          if (!g[svc]) g[svc] = [];
          g[svc].push(pod);
        }
      }
    }

    // Only subdivide when there are 2+ distinct nodes with real names
    const allKeys = [...nodeGroups.keys()];
    const hasRealNodes = allKeys.some(k => k !== '__unknown__');
    if (allKeys.length <= 1 || !hasRealNodes) return zones;

    // Sort: most-upstream service node first; break ties alphabetically
    allKeys.sort((a, b) => {
      const sa = nodeScore(nodeGroups.get(a));
      const sb = nodeScore(nodeGroups.get(b));
      if (sa !== sb) return sa - sb;
      return a.localeCompare(b);
    });

    return allKeys.map((nodeName, i) => ({
      zone:        zones[0]?.zone   || '',
      region:      zones[0]?.region || '',
      label:       `Node ${i + 1}`,
      icon:        baseIcon,
      _actualNode: nodeName !== '__unknown__' ? nodeName : null,
      pods:        nodeGroups.get(nodeName),
    }));
  }

  const opZonesFinal = subdivideByNode(opZones, '🖥️');
  const ewZonesFinal = subdivideByNode(ewZones, '🔗');

  // Render a labelled group section containing a row of zone cards
  function infraGroup(icon, title, zones, cardModFn) {
    if (!zones.length) return '';
    const cardsHtml = zones.map(zone => {
      const mod = cardModFn ? cardModFn(zone) : '';
      return renderInfraZoneCard(zone, svcOrder, mod);
    }).join('');
    return `
      <div class="infra-group">
        <div class="infra-group-header">
          <span class="infra-group-icon">${icon}</span>
          <span class="infra-group-name">${esc(title)}</span>
        </div>
        <div class="infra-zones-row">${cardsHtml}</div>
      </div>`;
  }

  const html =
    infraGroup('🌐', regionStr || 'Cloud', cloudZones,    () => '') +
    infraGroup('🔗', 'External Workloads', ewZonesFinal,  () => ' is-ew') +
    infraGroup('🏢', 'On-Premise',         opZonesFinal,  () => ' is-onprem');

  content.innerHTML = html;
}

// Renders one zone card — called by infraGroup() inside renderInfrastructure.
function renderInfraZoneCard(zone, svcOrder, cardMod) {
  let html = `<div class="infra-zone-card${cardMod}">`;

  const nodeTip = zone._actualNode ? ` data-tooltip="${esc('Node: ' + zone._actualNode)}"` : '';
  html += `<div class="infra-zone-header">
    <span class="infra-zone-icon">${zone.icon}</span>
    <span class="infra-zone-name"${nodeTip}>${esc(zone.label)}</span>
  </div>`;

  // Render services in the predefined order first, then any extras
  // (smiley2, smiley3, color2, color3 …) not in the fixed order list.
  function renderSvcGroup(svc, pods) {
    // Inherit icon/label from the base service name when the exact key is absent
    const baseSvc = Object.keys(INFRA_SVC_META).find(k => svc === k || svc.startsWith(k)) || svc;
    const baseMeta = INFRA_SVC_META[baseSvc] || { icon: '⬡', label: svc };
    const meta = INFRA_SVC_META[svc] || { icon: baseMeta.icon, label: svc };
    let h = `<div class="infra-svc-group">
      <div class="infra-svc-label">
        <span class="infra-svc-icon">${meta.icon}</span>
        <span class="infra-svc-name">${esc(meta.label)}</span>
      </div>`;
    for (const pod of pods) {
      const shortName  = shortPodName(pod.name || pod.ip);
      const stateEl    = infraPodStateEl(pod, baseSvc); // use base type for state rendering
      const tip        = infraPodTip(pod, baseSvc);
      const unavailCls = pod.available ? '' : ' infra-pod-unavail';
      const colorAttr  = (baseSvc === 'color' && pod.color) ? ` data-pod-color="${esc(pod.color)}"` : '';
      const colorEdgeAttr = (baseSvc === 'color' && pod.colorEdge) ? ` data-pod-color-edge="${esc(pod.colorEdge)}"` : '';
      h += `<div class="infra-pod-row${unavailCls}" data-tooltip="${esc(tip)}"${colorAttr}${colorEdgeAttr}>
        ${stateEl}
        <span class="infra-pod-name">${esc(shortName)}</span>
      </div>`;
    }
    h += `</div>`;
    return h;
  }

  const rendered = new Set();
  for (const svc of svcOrder) {
    const pods = zone.pods?.[svc];
    if (!pods?.length) continue;
    rendered.add(svc);
    html += renderSvcGroup(svc, pods);
  }
  // Render any additional services not in the fixed order (smiley2, color3 …)
  for (const [svc, pods] of Object.entries(zone.pods || {})) {
    if (rendered.has(svc) || !pods?.length) continue;
    html += renderSvcGroup(svc, pods);
  }

  html += `</div>`; // zone-card
  return html;
}

// Returns inline HTML showing the serving state for one pod.
function infraPodStateEl(pod, svc) {
  if (svc === 'smiley') {
    if (pod.smiley) {
      const glyph = v => {
        const src = smileyImgSrc(v);
        return src
          ? `<img class="infra-state-img" src="${esc(src)}" alt="">`
          : `<span class="infra-state-emoji">${decodeEntity(v)}</span>`;
      };
      const center = glyph(pod.smiley);
      // Edge differs → show both, side by side (tooltip says which is which)
      return pod.smileyEdge ? center + glyph(pod.smileyEdge) : center;
    }
  }
  if (svc === 'color') {
    if (pod.color) {
      const center = `<span class="infra-state-color" style="background:${esc(pod.color)}"></span>`;
      return pod.colorEdge
        ? center + `<span class="infra-state-color" style="background:${esc(pod.colorEdge)}"></span>`
        : center;
    }
  }
  if (svc === 'publisher') {
    if (!pod.available) return `<span class="infra-dot dot-red"></span>`;
    if (pod.paused) return `<span class="infra-state-rate infra-paused">⏸</span>`;
    const rate = pod.publishIntervalMs === 0 ? '∞' : msToRate(pod.publishIntervalMs);
    return `<span class="infra-state-rate">${esc(String(rate))}</span>`;
  }
  if (svc === 'subscriber') {
    if (!pod.available) return `<span class="infra-dot dot-red"></span>`;
    if (pod.paused)     return `<span class="infra-state-rate infra-paused">⏸</span>`;
    return `<span class="infra-dot dot-green"></span>`;
  }
  // face / gui — just a health dot
  return pod.available
    ? `<span class="infra-dot dot-green"></span>`
    : `<span class="infra-dot dot-red"></span>`;
}

// Returns multi-line tooltip text for one pod in the infra view.
function infraPodTip(pod, svc) {
  const lines = [];
  // Status leads the tooltip — no label when there's a single value. When center & edge
  // differ, both are shown labeled. Color shows ONLY swatch(es) — injected by the tooltip
  // renderer from data-pod-color / data-pod-color-edge; the hex value is omitted.
  if (svc === 'smiley' && pod.smiley) {
    const tipGlyph = v => smileyImgSrc(v) ? (smileyImgLabel(v) || 'custom image') : decodeEntity(v);
    if (pod.smileyEdge) {
      lines.push(`Center: ${tipGlyph(pod.smiley)}`);
      lines.push(`Edge: ${tipGlyph(pod.smileyEdge)}`);
    } else {
      lines.push(tipGlyph(pod.smiley));
    }
  }
  if (pod.name && pod.name !== pod.ip) lines.push(`Pod: ${pod.name}`);
  lines.push(`IP: ${pod.ip}`);
  if (pod.node)   lines.push(`Node: ${pod.node}`);
  if (pod.zone)   lines.push(`Zone: ${pod.zone}`);
  if (pod.region) lines.push(`Region: ${pod.region}`);
  if (svc === 'publisher') {
    if (pod.publishIntervalMs != null) {
      // msToRate() already includes the "msg/s" unit — don't append it again
      const r = pod.publishIntervalMs === 0
        ? '∞ (flood mode)'
        : msToRate(pod.publishIntervalMs);
      lines.push(`Rate: ${r}`);
    }
    lines.push(`Status: ${pod.paused ? '⏸ paused' : '▶ running'}`);
  }
  if (svc === 'subscriber') {
    lines.push(`Status: ${pod.paused ? '⏸ paused' : '▶ running'}`);
  }
  // Active fault injection values for this pod (per-pod chaos from /api/infrastructure)
  const cs = pod.chaos;
  if (cs) {
    const faults = [];
    if (cs.errorFraction > 0) faults.push(`  • Errors: ${cs.errorFraction}%`);
    const tipDelays = (cs.delayBuckets || []).filter(ms => ms > 0);
    if (tipDelays.length) faults.push(`  • Delay: ${tipDelays.join(', ')} ms`);
    if (cs.maxRate > 0) faults.push(`  • Max rate: ${cs.maxRate} RPS`);
    if (cs.latchFraction > 0) faults.push(`  • Latch chance: ${cs.latchFraction}%`);
    if (cs.latched) faults.push(`  • Latched into 599 (active)`);
    if (faults.length) lines.push('Faults:', ...faults);
  }
  if (!pod.available && pod.error) lines.push(`Error: ${pod.error}`);
  return lines.join('\n');
}

// ── Settings mode toggle ──────────────────────────────────────────────────
// Syncs the toggle button state from the current config, then wires click handlers.
function updateModeToggleButtons(currentMode) {
  document.querySelectorAll('#settings-mode-toggle .toggle-btn').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.mode === currentMode);
  });
}

function initModeToggle() {
  document.querySelectorAll('#settings-mode-toggle .toggle-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
      const newMode = btn.dataset.mode;
      try {
        const r = await fetch('/api/config', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ faceMode: newMode }),
        });
        if (r.ok) {
          showToast(`Mode switched to ${newMode === 'pubsub' ? 'Pub·Sub' : 'Classic'}`, 'success');
          poll(); // re-poll so all pages update immediately
        } else {
          showToast('Mode switch failed', 'error');
        }
      } catch (e) {
        showToast('Mode switch error: ' + e.message, 'error');
      }
    });
  });
}

// ── Emojivoto Integration ──────────────────────────────────────────────────
const EMOJIVOTO_POLL_MS = 5000;

function initEmojivoto() {
  const saveBtn = document.getElementById('ev-save-btn');
  if (saveBtn) saveBtn.addEventListener('click', saveEmojivotoConfig);
  const input = document.getElementById('ev-endpoint-input');
  if (input) input.addEventListener('keydown', e => { if (e.key === 'Enter') saveEmojivotoConfig(); });

  // Apply-To toggle — same pattern as Controls page
  document.querySelectorAll('#ev-apply-target-toggle .toggle-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('#ev-apply-target-toggle .toggle-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      evApplyTarget = btn.dataset.evTarget;
      evApplyDirty = true; // don't let the settings poll clobber this before Save
    });
  });
}

async function pollEmojivotoSettings() {
  try {
    const [data, pods, smileyState] = await Promise.all([
      fetchJSON('/api/emojivoto'),
      fetchJSON('/api/smileypods').catch(() => []),
      fetchJSON('/api/smileystate').catch(() => []),
    ]);
    lastEmojivotoData = data;
    renderEmojivotoStatus(data);
    if (data.leaderboard && data.leaderboard.length > 0) {
      renderEmojivotoLeaderboard(data.leaderboard);
    }
    const smileyStateMap = Object.fromEntries((smileyState || []).map(s => [s.ip, s]));
    renderEmojivotoSmileyPods(pods, data.selectedPods || [], smileyStateMap);
    if (state.lastConfig) renderPollTooltip(state.lastConfig);
  } catch (_) {}
}

function renderEmojivotoStatus(data) {
  const chip        = document.getElementById('ev-status-chip');
  const leaderWrap  = document.getElementById('ev-leader-wrap');
  const leaderEmoji = document.getElementById('ev-leader-emoji');
  const lbWrap      = document.getElementById('ev-leaderboard-wrap');
  const podsSection = document.getElementById('ev-pods-section');
  const endpointInput = document.getElementById('ev-endpoint-input');
  const enabledCheck  = document.getElementById('ev-enabled-check');

  const updateSmileyCheck = document.getElementById('ev-update-smileys-check');

  if (endpointInput && !endpointInput.matches(':focus')) endpointInput.value = data.endpoint || '';
  if (enabledCheck) enabledCheck.checked = data.enabled || false;
  if (updateSmileyCheck) updateSmileyCheck.checked = data.updateSmileys || false;

  // Restore the Apply-To toggle from the persisted server state — unless the
  // user has an unsaved selection pending.
  const which = data.which || 'all';
  if (!evApplyDirty && which !== evApplyTarget) {
    evApplyTarget = which;
    document.querySelectorAll('#ev-apply-target-toggle .toggle-btn').forEach(b => {
      b.classList.toggle('active', b.dataset.evTarget === which);
    });
  }

  if (chip) {
    chip.dataset.status = data.status;
    if (data.status === 'ok')           chip.textContent = '● Connected';
    else if (data.status === 'error')   chip.textContent = '✕ ' + (data.error || 'error');
    else                                chip.textContent = 'Not configured';
  }

  const evActive = data.enabled && data.endpoint;
  if (lbWrap)      lbWrap.style.display      = (data.status === 'ok' && data.leaderboard && data.leaderboard.length > 0) ? '' : 'none';
  if (podsSection) podsSection.style.display = evActive ? '' : 'none';
  const targetSection = document.getElementById('ev-target-section');
  if (targetSection) targetSection.style.display = evActive ? '' : 'none';
  if (updateSmileyCheck) updateSmileyCheck.disabled = !(data.enabled && data.endpoint);

  if (leaderWrap) {
    if (data.leader) {
      leaderWrap.style.display = '';
      if (leaderEmoji) leaderEmoji.textContent = data.leader;
    } else {
      leaderWrap.style.display = 'none';
    }
  }
}

function renderEmojivotoLeaderboard(entries) {
  const tbody = document.getElementById('ev-leaderboard-tbody');
  if (!tbody) return;
  tbody.innerHTML = entries.slice(0, 10).map((e, i) =>
    `<tr>
      <td style="color:var(--text-muted);width:28px">${i + 1}</td>
      <td style="font-size:18px;line-height:1;width:32px">${e.unicode}</td>
      <td style="text-align:right;font-variant-numeric:tabular-nums;font-family:monospace">${esc(e.votes)}</td>
    </tr>`
  ).join('');
}

function renderEmojivotoSmileyPods(pods, serverSelectedPods, stateMap = {}) {
  // On first render, seed selection from server state
  if (selectedEmojivotoPods.length === 0 && serverSelectedPods && serverSelectedPods.length > 0) {
    selectedEmojivotoPods = [...serverSelectedPods];
  }
  renderServicePods('ev-pod-sel-section', 'ev-pod-sel-row', pods, 'emojivoto', stateMap);
}

async function saveEmojivotoConfig() {
  const inputEl      = document.getElementById('ev-endpoint-input');
  const endpoint     = (inputEl.value.trim() || inputEl.placeholder || '').trim();
  const enabled      = document.getElementById('ev-enabled-check').checked;
  const updateSmileys = document.getElementById('ev-update-smileys-check').checked;
  const btn      = document.getElementById('ev-save-btn');
  const statusEl = document.getElementById('ev-save-status');

  btn.disabled = true;
  try {
    const r = await fetch('/api/emojivoto', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ endpoint, enabled, updateSmileys, which: evApplyTarget, selectedPods: selectedEmojivotoPods }),
    });
    if (!r.ok) throw new Error(`HTTP ${r.status}`);
    evApplyDirty = false;
    if (statusEl) { statusEl.textContent = 'Saved'; setTimeout(() => { statusEl.textContent = ''; }, 2000); }
    await pollEmojivotoSettings();
  } catch (e) {
    if (statusEl) statusEl.textContent = 'Error saving';
  } finally {
    btn.disabled = false;
  }
}

// ── Boot ───────────────────────────────────────────────────────────────────
initTheme();
initTooltips();
initSidebar();
initMaintenance();
initEndpointEditor();
initModeToggle();
initLiveView();
initSmileyPicker();
initColorPicker();
initApplyControls();
initScenarioButtons();
initEmojivoto();
initChaos();
initFaultInjection();
poll();
pollEmojivotoSettings();
setInterval(poll, POLL_INTERVAL);
pollInfra();
setInterval(pollInfra, INFRA_POLL_MS);
setInterval(pollEmojivotoSettings, EMOJIVOTO_POLL_MS);
window.addEventListener('resize', () => { drawCharts(); sizeLiveGrid(); });
new MutationObserver(() => drawCharts())
  .observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });

// ── Chaos injection ────────────────────────────────────────────────────────

// Services to show in classic vs pubsub mode.
const CHAOS_SERVICES_CLASSIC = ['smiley', 'color', 'face'];
const CHAOS_SERVICES_PUBSUB  = ['smiley', 'color', 'publisher', 'subscriber'];

// Labels for display
const CHAOS_SERVICE_LABELS = {
  smiley: 'Smiley', color: 'Color', face: 'Face',
  publisher: 'Publisher', subscriber: 'Subscriber',
};

function initChaos() {
  const header = document.getElementById('chaos-toggle');
  const body   = document.getElementById('chaos-body');
  const chevron = document.getElementById('chaos-chevron');
  if (!header || !body) return;

  const toggle = () => {
    const open = body.style.display !== 'none';
    body.style.display = open ? 'none' : 'block';
    chevron.textContent = open ? '▸' : '▾';
    header.setAttribute('aria-expanded', String(!open));
  };
  header.addEventListener('click', toggle);
  header.addEventListener('keydown', e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggle(); } });
}

function renderChaos(chaos, mode) {
  const container = document.getElementById('chaos-cards');
  if (!container) return;

  const services = (mode === 'classic' ? CHAOS_SERVICES_CLASSIC : CHAOS_SERVICES_PUBSUB);

  // Build cards — reuse existing DOM nodes keyed by service name to avoid flicker.
  services.forEach(svc => {
    let card = container.querySelector(`.chaos-card[data-svc="${svc}"]`);
    const state = chaos[svc] || { available: false };

    if (!card) {
      card = document.createElement('div');
      card.className = 'chaos-card';
      card.dataset.svc = svc;
      card.innerHTML = buildChaosCardHTML(svc);
      container.appendChild(card);
      wireChaosCard(card, svc);
    }

    updateChaosCard(card, svc, state);
  });

  // Remove cards for services not in the current mode
  container.querySelectorAll('.chaos-card').forEach(card => {
    if (!services.includes(card.dataset.svc)) card.remove();
  });
}

function buildChaosCardHTML(svc) {
  const label = CHAOS_SERVICE_LABELS[svc] || svc;
  return `
    <div class="chaos-card-header">
      <span class="chaos-svc-name">${label}</span>
      <span class="chaos-latched-badge" style="display:none">⚠ LATCHED</span>
    </div>
    <div class="chaos-field-row">
      <label class="chaos-label">Error %</label>
      <input type="range" class="chaos-slider" data-field="errorFraction" min="0" max="100" value="0">
      <span class="chaos-slider-val">0</span>
    </div>
    <div class="chaos-field-row">
      <label class="chaos-label">Latch %</label>
      <input type="range" class="chaos-slider" data-field="latchFraction" min="0" max="100" value="0">
      <span class="chaos-slider-val">0</span>
    </div>
    <div class="chaos-field-row">
      <label class="chaos-label">Delays (ms)</label>
      <input type="text" class="chaos-text-input" data-field="delayBuckets" placeholder="e.g. 100,500,1000">
    </div>
    <div class="chaos-field-row">
      <label class="chaos-label">Max Rate (RPS)</label>
      <input type="number" class="chaos-text-input chaos-rate-input" data-field="maxRate" placeholder="0 = off" min="0" step="0.5">
    </div>
    <div class="chaos-card-actions">
      <button class="chaos-btn chaos-apply-btn primary">Apply</button>
      <button class="chaos-btn chaos-reset-btn">Reset</button>
      <button class="chaos-btn chaos-unlatch-btn danger" style="display:none">Force Unlatch</button>
    </div>
    <div class="chaos-card-status"></div>
  `;
}

function wireChaosCard(card, svc) {
  // Sync slider → readout
  card.querySelectorAll('.chaos-slider').forEach(slider => {
    const readout = slider.nextElementSibling;
    slider.addEventListener('input', () => { readout.textContent = slider.value; });
  });

  card.querySelector('.chaos-apply-btn').addEventListener('click', () => applyChaos(card, svc));
  card.querySelector('.chaos-reset-btn').addEventListener('click', () => resetChaos(card, svc));
  card.querySelector('.chaos-unlatch-btn').addEventListener('click', () => forceUnlatch(card, svc));
}

function updateChaosCard(card, svc, state) {
  if (!state.available && state.error) {
    card.querySelector('.chaos-card-status').textContent = '⚠ ' + state.error;
    return;
  }

  // Only update sliders/fields if user hasn't interacted (no :focus)
  const focused = card.querySelector(':focus');
  if (!focused) {
    setSlider(card, 'errorFraction', state.errorFraction || 0);
    setSlider(card, 'latchFraction', state.latchFraction || 0);
    const db = card.querySelector('[data-field="delayBuckets"]');
    if (db && document.activeElement !== db)
      db.value = (state.delayBuckets || []).join(',');
    const mr = card.querySelector('[data-field="maxRate"]');
    if (mr && document.activeElement !== mr)
      mr.value = state.maxRate || '';
  }

  const latchBadge  = card.querySelector('.chaos-latched-badge');
  const unlatchBtn  = card.querySelector('.chaos-unlatch-btn');
  const isLatched   = !!state.latched;
  latchBadge.style.display  = isLatched ? '' : 'none';
  unlatchBtn.style.display  = isLatched ? '' : 'none';
}

function setSlider(card, field, value) {
  const slider  = card.querySelector(`[data-field="${field}"]`);
  const readout = slider && slider.nextElementSibling;
  if (slider && document.activeElement !== slider) {
    slider.value = value;
    if (readout) readout.textContent = value;
  }
}

// chaosPodStatus interprets the {pods, succeeded, failed} response body from
// PUT /api/chaos/{svc}. The backend returns HTTP 200 even on pod-level failures,
// so callers must inspect succeeded vs failed rather than relying on r.ok alone.
function chaosPodStatus(data) {
  const succeeded = data.succeeded ?? (data.pods ?? 1);
  const total     = data.pods     ?? 1;
  const failed    = data.failed   ?? 0;
  if (succeeded === 0)
    return { text: `✗ No pods reached (${total} attempted)`, isError: true, isWarn: false };
  if (failed > 0)
    return { text: `⚠ Applied to ${succeeded}/${total} pod(s)`, isError: false, isWarn: true };
  return { text: `✓ Applied to ${succeeded} pod(s)`, isError: false, isWarn: false };
}

async function applyChaos(card, svc) {
  const body = {};
  const ef = card.querySelector('[data-field="errorFraction"]');
  if (ef) body.errorFraction = parseInt(ef.value, 10);
  const lf = card.querySelector('[data-field="latchFraction"]');
  if (lf) body.latchFraction = parseInt(lf.value, 10);
  const db = card.querySelector('[data-field="delayBuckets"]');
  if (db) {
    const raw = db.value.trim();
    // Zeros are only meaningful alongside real delays ("0,500" = half the
    // requests undelayed); a zeros-only list is just "no delay" and would
    // paint a phantom ⏱ badge, so it's sent as empty.
    const nums = raw ? raw.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !isNaN(n) && n >= 0) : [];
    body.delayBuckets = nums.some(n => n > 0) ? nums : [];
  }
  const mr = card.querySelector('[data-field="maxRate"]');
  if (mr && mr.value !== '') body.maxRate = parseFloat(mr.value);

  const statusEl = card.querySelector('.chaos-card-status');
  const applyBtn = card.querySelector('.chaos-apply-btn');
  applyBtn.disabled = true;
  statusEl.textContent = 'Applying…';

  try {
    const r = await fetch(`/api/chaos/${svc}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const data = await r.json();
    if (r.ok) {
      const ps = chaosPodStatus(data);
      statusEl.textContent = ps.text;
      if (!ps.isError) setTimeout(() => { statusEl.textContent = ''; }, 3000);
    } else {
      statusEl.textContent = '✗ ' + (data.error || r.statusText);
    }
  } catch (e) {
    statusEl.textContent = '✗ ' + e.message;
  } finally {
    applyBtn.disabled = false;
  }
}

async function resetChaos(card, svc) {
  const body = { errorFraction: 0, latchFraction: 0, delayBuckets: [], maxRate: 0, forceUnlatch: true };
  const statusEl = card.querySelector('.chaos-card-status');
  statusEl.textContent = 'Resetting…';
  try {
    const r = await fetch(`/api/chaos/${svc}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    const data = await r.json();
    if (r.ok) {
      const ps = chaosPodStatus(data);
      statusEl.textContent = ps.text.replace('Applied', 'Reset');
      if (!ps.isError) setTimeout(() => { statusEl.textContent = ''; }, 3000);
    } else {
      statusEl.textContent = '✗ ' + (data.error || r.statusText);
    }
  } catch (e) {
    statusEl.textContent = '✗ ' + e.message;
  }
}

async function forceUnlatch(card, svc) {
  const statusEl = card.querySelector('.chaos-card-status');
  statusEl.textContent = 'Unlatching…';
  try {
    const r = await fetch(`/api/chaos/${svc}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ forceUnlatch: true }),
    });
    if (r.ok) {
      statusEl.textContent = '✓ Unlatched';
      setTimeout(() => { statusEl.textContent = ''; }, 3000);
    } else {
      const d = await r.json().catch(() => ({}));
      statusEl.textContent = '✗ ' + (d.error || r.statusText);
    }
  } catch (e) {
    statusEl.textContent = '✗ ' + e.message;
  }
}

// ── Fault Injection page ────────────────────────────────────────────────────

const FI_EFFECTS = {
  smiley:     { classic: 'Errors return HTTP 500 → cells show cursing face 🤬. Delays slow per-cell refresh.',
                pubsub:  'Publisher calls smiley each cycle. Errors inject bad content into MySQL → queue fills with error faces.' },
  color:      { classic: 'Errors return HTTP 500 → cells lose color (grey fallback). Delays add gRPC latency per cell.',
                pubsub:  'Publisher calls color via gRPC. Errors inject grey-background faces into MySQL → GUI receives unstyled content.' },
  face:       { classic: 'Face orchestrates smiley + color. Errors take down every GUI cell at once — the entire grid goes dark.',
                pubsub:  'Not active in pub/sub mode — the face alias points at face-subscriber.' },
  publisher:  { pubsub:  'Errors reduce queue fill rate. The queue drains faster than it fills → GUI cells eventually show stale or no faces. Latching halts publishing entirely for 30 s.' },
  subscriber: { pubsub:  'Errors cause GUI cell fetches to fail. Max RPS caps cells served per second. Latching puts subscriber in 599 state — all cells fail until unlatched.' },
};
const FI_ICONS   = { smiley: '😃', color: '🎨', face: '😀', publisher: '📤', subscriber: '📥' };
const FI_BANNERS = {};

// Preset delay values available as clickable bubbles
const FI_DELAY_PRESETS = [50, 100, 200, 500, 1000, 2000];

// Per-service pod selection state (IP arrays, keyed by service name)
const fiSelectedPods = {};   // { smiley: [], color: [], … }
const fiPodCache     = {};   // { smiley: [PodTopology…], … }
// Endpoint that returns pod list for each service
const FI_PODS_ENDPOINT = {
  smiley: '/api/smileypods', color: '/api/colorpods', face: '/api/facepods',
  // publisher/subscriber: pods come from state.lastControls (already polled every 3s)
  publisher: '__controls__', subscriber: '__controls__',
};

// initFaultInjection is a no-op at boot — the scenario bar and global panel are
// built mode-aware inside renderFaultInjection (first render / on mode change),
// which also wires their event handlers via buildFIGlobalPanel/renderFIScenarioBar.
function initFaultInjection() { /* built lazily in renderFaultInjection */ }

// ── Predefined scenarios ────────────────────────────────────────────────────
// Each scenario applies its body to all pods of its services (scope ignored).
const FI_SCENARIOS = [
  { key: 'recover',      label: '↺ Recover All', dataScenario: 'recover',
    mode: 'both',    servicesFor: m => (m === 'classic' ? CHAOS_SERVICES_CLASSIC : CHAOS_SERVICES_PUBSUB),
    body: { errorFraction: 0, latchFraction: 0, maxRate: 0, delayBuckets: [], forceUnlatch: true },
    note: 'Clears all fault injection on every service.' },
  { key: 'smiley-errors', label: '🤬 Smiley Errors', dataScenario: 'errors',
    mode: 'both',    servicesFor: () => ['smiley'],
    body: { errorFraction: 75, latchFraction: 0, maxRate: 0, delayBuckets: [] },
    note: '75% of smiley requests return errors → cursing faces.' },
  { key: 'color-latency', label: '🐌 Color Latency', dataScenario: 'latency',
    mode: 'both',    servicesFor: () => ['color'],
    body: { errorFraction: 0, latchFraction: 0, maxRate: 0, delayBuckets: [500, 1000, 2000] },
    note: 'Adds 500–2000ms random delay to every color request.' },
  { key: 'rate-limit',   label: '🚦 Rate Limit', dataScenario: 'throttle',
    mode: 'both',    servicesFor: m => (m === 'classic' ? ['face'] : ['subscriber']),
    body: { errorFraction: 0, latchFraction: 0, maxRate: 2, delayBuckets: [] },
    note: 'Caps the request rate at 2 RPS → 429s above that.' },
  { key: 'smiley-latch', label: '🔒 Latch Smiley', dataScenario: 'errors',
    mode: 'both',    servicesFor: () => ['smiley'],
    body: { errorFraction: 50, latchFraction: 80, maxRate: 0, delayBuckets: [] },
    note: 'Errors latch smiley into persistent 599 state for 30s.' },
];

// Tracks which mode the scenario bar + global panel were last built for.
let fiBuiltMode = null;

function renderFIScenarioBar(mode) {
  const bar = document.getElementById('fi-scenario-bar');
  if (!bar) return;
  // Clear existing buttons (keep the label)
  bar.querySelectorAll('.scenario-btn').forEach(b => b.remove());
  FI_SCENARIOS.filter(s => s.mode === 'both' || s.mode === mode).forEach(s => {
    const btn = document.createElement('button');
    btn.className = 'scenario-btn';
    btn.dataset.scenario = s.dataScenario;
    btn.dataset.key = s.key;
    btn.textContent = s.label;
    btn.title = s.note;
    btn.addEventListener('click', () => runFIScenario(s.key));
    bar.appendChild(btn);
  });
}

function setFIScenarioDisabled(disabled) {
  document.querySelectorAll('#fi-scenario-bar .scenario-btn').forEach(b => { b.disabled = disabled; });
}

async function runFIScenario(key) {
  const sc = FI_SCENARIOS.find(s => s.key === key);
  if (!sc) return;
  const services = sc.servicesFor(state.mode);
  setFIScenarioDisabled(true);
  // Scenarios always target all pods — scope is ignored.
  const res = await applyGlobalChaos(sc.body, services, 'all', null);
  setFIScenarioDisabled(false);
  refreshFITopologyBadges();
  showToast(`⚡ ${sc.label.replace(/^[^\sA-Za-z]+\s*/, '')} — ${res.summary}`, res.failed > 0 ? 'error' : 'success');
}

// ── Dynamic zone / node scenario rows ────────────────────────────────────────
// Builds Outage + Latency rows from the live topology — one button per cloud
// zone and per on-prem/EW node. Hidden until topology data is available.
// All rows live inside the unified .fi-scenario-panel card.
function renderFIOutageButtons(mode) {
  const section = document.getElementById('fi-outage-section');
  if (!section) return;

  const infra = state.lastInfra;
  if (!infra || !infra.zones || !infra.zones.length) {
    section.style.display = 'none';
    return;
  }

  const { groups } = fiBuildTopologyGroups(infra);
  const svcs = mode === 'classic' ? CHAOS_SERVICES_CLASSIC : CHAOS_SERVICES_PUBSUB;

  const zoneCards = [], ewCards = [], nodeCards = [];
  for (const g of groups) {
    for (const c of g.cards) {
      if (c.scope === 'zone')                    zoneCards.push(c);
      else if (g.title === 'External Workloads') ewCards.push({ ...c, groupTitle: g.title });
      else                                       nodeCards.push({ ...c, groupTitle: g.title });
    }
  }

  if (!zoneCards.length && !ewCards.length && !nodeCards.length) {
    section.style.display = 'none';
    return;
  }

  // Helper: build a .fi-scenario-row with a label + outage button + latency button per card.
  function scopeRow(rowLabel, cards, scopeHint) {
    const outage  = { errFraction: 100, label: 'Outage', icon: '💥', type: 'outage',   btnCls: 'fi-outage-btn',  scenario: 'outage'  };
    const latency = { errFraction: 0,   label: 'Latency', icon: '⏱',  type: 'latency', btnCls: 'fi-latency-btn', scenario: 'latency' };
    let html = '';
    for (const kind of [outage, latency]) {
      html += `<div class="fi-scenario-row">
        <span class="scenario-label">${esc(rowLabel)} ${esc(kind.label)}</span>`;
      for (const c of cards) {
        const nodeTip = c.actualNode ? ` (node: ${c.actualNode})` : '';
        const tip = `${kind.label === 'Outage' ? '100% errors' : '500–2000 ms delays'} on all pods in ${c.label}${nodeTip}`;
        html += `<button class="scenario-btn fi-scope-scenario-btn ${kind.btnCls}"
          data-scenario="${kind.scenario}" data-scenario-type="${kind.type}"
          data-scope="${esc(c.scope)}" data-detail="${esc(c.detail || '')}"
          title="${esc(tip)}">${kind.icon} ${esc(c.label)}</button>`;
      }
      html += `</div>`;
    }
    return html;
  }

  let html = '';
  if (zoneCards.length) html += scopeRow('Zone',         zoneCards, 'zone');
  if (ewCards.length)   html += scopeRow('Ext Workload', ewCards,   'node');
  if (nodeCards.length) html += scopeRow('Node',         nodeCards, 'node');

  section.innerHTML = html;
  section.style.display = '';

  const OUTAGE_BODY  = { errorFraction: 100, latchFraction: 0, maxRate: 0, delayBuckets: [] };
  const LATENCY_BODY = { errorFraction: 0,   latchFraction: 0, maxRate: 0, delayBuckets: [500, 1000, 2000] };

  section.querySelectorAll('.fi-scope-scenario-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const body  = btn.dataset.scenarioType === 'latency' ? LATENCY_BODY : OUTAGE_BODY;
      const icon  = btn.dataset.scenarioType === 'latency' ? '⏱' : '💥';
      const label = `${icon} ${btn.textContent.trim()}`;
      runFIScopeScenario(btn.dataset.scope, btn.dataset.detail || null, label, body, svcs);
    });
  });
}

async function runFIScopeScenario(scope, detail, label, body, svcs) {
  document.querySelectorAll('.fi-scope-scenario-btn').forEach(b => { b.disabled = true; });
  const res = await applyGlobalChaos(body, svcs, scope, detail || null);
  document.querySelectorAll('.fi-scope-scenario-btn').forEach(b => { b.disabled = false; });
  refreshFITopologyBadges();

  // Detect total failure (all pods unreachable) — likely cross-node direct IP connectivity issue
  const allSkipped = res.skipped && res.skipped.length === svcs.length;
  const allFailed  = res.failed > 0 && !res.summary.includes('applied to 0') && !allSkipped;
  const kind = allSkipped ? 'warn'
    : allFailed ? 'error'
    : 'success';
  const suffix = allSkipped
    ? ' — no pods matched this scope'
    : allFailed
    ? ' — pods unreachable (may be on a different node)'
    : '';
  showToast(`${label}${suffix} — ${res.summary}`, kind);
}

// ── Scope → IP resolution ─────────────────────────────────────────────────
// Map an infra service key (smiley2, color3 …) to its base chaos service name.
function chaosBaseSvc(svcKey) {
  for (const base of ['smiley', 'color', 'face', 'publisher', 'subscriber', 'gui']) {
    if (svcKey === base || svcKey.startsWith(base)) return base;
  }
  return svcKey;
}

// Resolve {svc: [ip…]} for a given scope across the selected services.
function resolveScopeTargets(scope, detailKey, services) {
  const targets = {};
  const infra = state.lastInfra;
  if (!infra || !infra.zones) return targets;
  const svcSet = new Set(services);
  for (const z of infra.zones) {
    for (const [svcKey, pods] of Object.entries(z.pods || {})) {
      const base = chaosBaseSvc(svcKey);
      if (!svcSet.has(base)) continue;
      for (const p of pods) {
        const ip = p.ip || p.IP;
        if (!ip) continue;
        let match = false;
        switch (scope) {
          case 'zone':     match = p.zone === detailKey; break;
          case 'node':     match = p.node === detailKey; break;
          case 'ewip':     match = (p.ip || p.IP) === detailKey; break;
          case 'external': match = p.workloadType === 'externalworkload'; break;
          case 'onprem':   match = !p.zone && p.workloadType !== 'externalworkload'; break;
          default:         match = true;
        }
        if (match) {
          if (!targets[base]) targets[base] = [];
          if (!targets[base].includes(ip)) targets[base].push(ip);
        }
      }
    }
  }
  return targets;
}

// Core fan-out — the single place the global panel + scenarios PUT chaos.
// Returns { summary, failed } for status display.
async function applyGlobalChaos(body, services, scope, detail) {
  const skipped = [];
  let totalPods = 0, okServices = 0, failed = 0;

  const calls = [];
  if (scope === 'all') {
    // No pods field — backend hits all replicas of each service.
    for (const svc of services) {
      calls.push({ svc, body: { ...body } });
    }
  } else {
    const targets = resolveScopeTargets(scope, detail, services);
    for (const svc of services) {
      const ips = targets[svc];
      if (!ips || ips.length === 0) { skipped.push(svc); continue; }
      totalPods += ips.length;
      calls.push({ svc, body: { ...body, pods: ips } });
    }
  }

  const results = await Promise.allSettled(calls.map(c =>
    fetch(`/api/chaos/${c.svc}`, {
      method: 'PUT', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(c.body),
    }).then(r => r.json().then(d => ({ ok: r.ok, d })))
  ));

  results.forEach((r, i) => {
    if (r.status === 'fulfilled' && r.value.ok) {
      const d = r.value.d;
      // Treat as failed if ALL pods were unreachable (succeeded:0 with pods attempted)
      if (d && typeof d.succeeded === 'number' && d.succeeded === 0 && (d.pods ?? 0) > 0) {
        failed++;
      } else {
        okServices++;
        if (scope === 'all') totalPods += (d?.pods ?? 0);
      }
    } else {
      failed++;
    }
  });

  let summary = scope === 'all'
    ? `applied to ${okServices} service${okServices !== 1 ? 's' : ''}`
    : `applied to ${totalPods} pod${totalPods !== 1 ? 's' : ''} across ${okServices} service${okServices !== 1 ? 's' : ''}`;
  if (skipped.length) summary += ` (skipped: ${skipped.join(', ')} — no pods in scope)`;
  if (failed)         summary += ` · ${failed} failed`;
  return { summary, failed, skipped };
}

// ── Global apply panel ──────────────────────────────────────────────────────
function buildFIGlobalPanel(mode) {
  const panel = document.getElementById('fi-global-panel');
  if (!panel) return;
  const svcs = mode === 'classic' ? CHAOS_SERVICES_CLASSIC : CHAOS_SERVICES_PUBSUB;

  const svcPills = svcs.map(svc =>
    `<button class="fi-svc-pill active" data-svc="${svc}">${esc(CHAOS_SERVICE_LABELS[svc] || svc)}</button>`
  ).join('');

  const delayBubbles = FI_DELAY_PRESETS.map(ms =>
    `<button class="fi-delay-bubble" data-ms="${ms}" type="button">${ms}ms</button>`
  ).join('');

  panel.innerHTML = `
    <div class="fi-global-row">
      <span class="fi-global-label">Targeting</span>
      <span class="fi-target-label" id="fi-target-label">All pods</span>
      <span class="fi-target-count" id="fi-target-count"></span>
      <button class="fi-target-reset" id="fi-target-reset" type="button" title="Target all pods">↺ All pods</button>
    </div>

    <div class="fi-global-row">
      <span class="fi-global-label">Services</span>
      <div class="fi-svc-select" id="fi-svc-select">${svcPills}</div>
    </div>

    <div class="fi-param-group">
      <div class="fi-param-row">
        <div class="fi-param-label-row">
          <span class="fi-param-label">Error %</span>
          <span class="fi-slider-val" id="fi-g-err-val">Off</span>
        </div>
        <input type="range" class="fi-slider" data-field="errorFraction" min="0" max="100" value="0">
      </div>
      <div class="fi-param-row">
        <div class="fi-param-label-row">
          <span class="fi-param-label">Latch %</span>
          <span class="fi-slider-val" id="fi-g-latch-val">Off</span>
        </div>
        <input type="range" class="fi-slider" data-field="latchFraction" min="0" max="100" value="0">
      </div>
      <div class="fi-param-row">
        <div class="fi-param-label-row">
          <span class="fi-param-label">Max RPS</span>
          <span class="fi-slider-val" id="fi-g-rps-val">Off</span>
        </div>
        <input type="range" class="fi-slider" data-field="maxRate" min="0" max="20" step="0.5" value="0">
      </div>
      <div class="fi-param-row">
        <span class="fi-param-label">Delay (ms)</span>
        <div class="fi-delay-bubbles" data-field="delayBuckets">${delayBubbles}</div>
      </div>
    </div>

    <div class="fi-global-actions">
      <button class="fi-btn fi-apply-btn" id="fi-global-apply" type="button">Apply</button>
      <button class="fi-btn fi-reset-btn" id="fi-global-reset" type="button">Reset Selected</button>
      <button class="fi-btn fi-unlatch-btn" id="fi-global-unlatch" type="button">Force Unlatch</button>
      <span class="fi-global-status" id="fi-global-status"></span>
    </div>`;

  // Reset scope to all pods (scope itself is chosen by clicking topology cards)
  document.getElementById('fi-target-reset').addEventListener('click', () => {
    setFIScope('all', null);
  });

  // Service pills
  panel.querySelectorAll('#fi-svc-select .fi-svc-pill').forEach(pill => {
    pill.addEventListener('click', () => {
      pill.classList.toggle('active');
      state.fiSelectedServices = [...panel.querySelectorAll('#fi-svc-select .fi-svc-pill.active')].map(p => p.dataset.svc);
      updateFIEffectPreview();
    });
  });

  // Sliders → readout
  panel.querySelectorAll('.fi-slider').forEach(slider => {
    slider.addEventListener('input', () => updateFIGlobalSliderVal(slider.dataset.field, slider.value));
  });

  // Delay bubbles
  panel.querySelectorAll('.fi-delay-bubble').forEach(btn => {
    btn.addEventListener('click', () => btn.classList.toggle('active'));
  });

  // Actions
  document.getElementById('fi-global-apply').addEventListener('click', () => applyFIGlobal());
  document.getElementById('fi-global-reset').addEventListener('click', () => resetSelectedFI());
  document.getElementById('fi-global-unlatch').addEventListener('click', () => forceUnlatchSelectedFI());

  updateFITargetLabel();
  updateFIEffectPreview();
}

// Human-readable label for the current scope selection.
function fiScopeLabel() {
  switch (state.fiScope) {
    case 'zone':     return `Zone · ${state.fiScopeDetail}`;
    case 'node':     return `Node · ${shortPodName(state.fiScopeDetail || '')}`;
    case 'ewip':     return `EW · ${shortPodName(state.fiScopeDetail || '')}`;
    case 'external': return 'External Workloads';
    case 'onprem':   return 'On-Premise';
    default:         return 'All pods';
  }
}

// Set the active scope, refresh the topology highlight + targeting label + preview.
function setFIScope(scope, detail) {
  state.fiScope = scope;
  state.fiScopeDetail = detail || null;
  updateFITargetLabel();
  updateFIEffectPreview();
  state._fiTopoSig = null;               // force topology repaint so highlight updates
  if (state.lastInfra) renderFITopology(state.lastInfra, state.lastChaos || {}, state.mode);
}

function updateFITargetLabel() {
  const lbl = document.getElementById('fi-target-label');
  if (lbl) lbl.textContent = fiScopeLabel();
  const reset = document.getElementById('fi-target-reset');
  if (reset) reset.style.display = state.fiScope === 'all' ? 'none' : '';

  // Pod count — resolve targets and show how many pods will be affected.
  const countEl = document.getElementById('fi-target-count');
  if (!countEl) return;
  if (state.fiScope === 'all' || !state.lastInfra) {
    countEl.textContent = '';
    return;
  }
  const services = state.fiSelectedServices || [];
  const targets = resolveScopeTargets(state.fiScope, state.fiScopeDetail, services);
  const total = Object.values(targets).reduce((n, ips) => n + ips.length, 0);
  countEl.textContent = total === 0
    ? '(no pods in scope)'
    : `(${total} pod${total !== 1 ? 's' : ''})`;
}

function updateFIGlobalSliderVal(field, value) {
  const num = parseFloat(value);
  const map = { errorFraction: 'fi-g-err-val', latchFraction: 'fi-g-latch-val', maxRate: 'fi-g-rps-val' };
  const el = document.getElementById(map[field]);
  if (!el) return;
  if (num === 0) el.textContent = 'Off';
  else if (field === 'maxRate') el.textContent = `${num} RPS`;
  else el.textContent = `${num}%`;
}

// ── Status & Targeting topology (infra-style, clickable) ───────────────────
// Mirrors the Overview Infrastructure view: cloud zones, External Workloads,
// and On-Premise (node-subdivided). Each card is a clickable scope target, and
// every pod row shows its current fault status. Only existing groups render.

// Fault-status badges for a pod, derived from its service's chaos state.
function fiPodFaultBadges(baseSvc, chaos, pod) {
  // Per-pod chaos (from /api/infrastructure) is the source of truth — accurate for
  // node/zone scoped applies. The service-level aggregate is ONLY trustworthy for a
  // single-instance service; for multi-instance services it would paint a borrowed
  // fault on non-targeted pods, so we never use it there.
  let cs = (pod && pod.chaos) ? pod.chaos : null;
  if (!cs && !(state._fiMultiBases && state._fiMultiBases.has(baseSvc))) {
    cs = chaos[baseSvc]; // single-instance service → aggregate == this pod
  }
  if (!cs || cs.available === false) {
    // No per-pod data on a multi-instance service → show serving state, never a borrowed fault.
    return pod ? infraPodStateEl(pod, baseSvc) : `<span class="fi-fault-unknown" title="status unknown">?</span>`;
  }
  const badges = [];
  if (cs.errorFraction > 0) badges.push(`<span class="fi-fault-badge err" title="${cs.errorFraction}% errors">E ${cs.errorFraction}%</span>`);
  // Only real delays — a stored 0ms bucket is a no-op, not a fault
  const delays = (cs.delayBuckets || []).filter(ms => ms > 0);
  if (delays.length) badges.push(`<span class="fi-fault-badge delay" title="delays: ${delays.join(', ')}ms">⏱</span>`);
  if (cs.maxRate > 0) badges.push(`<span class="fi-fault-badge rate" title="max ${cs.maxRate} RPS">🚦${cs.maxRate}</span>`);
  if (cs.latched)     badges.push(`<span class="fi-fault-badge latch" title="latched into 599">🔒</span>`);
  // No active faults — show the actual serving value (emoji or colour swatch) from infra data.
  if (!badges.length) return pod ? infraPodStateEl(pod, baseSvc) : `<span class="fi-fault-ok" title="no faults">●</span>`;
  return badges.join('');
}

// Build grouped, node-subdivided cards from infra — mirrors renderInfrastructure
// but tags each card with a scope+detail for click-to-target. Self-contained so
// the Overview view is never touched.
function fiBuildTopologyGroups(infra) {
  const svcOrder = INFRA_SVC_ORDER[infra.mode] || INFRA_SVC_ORDER.classic;
  const SVC_NODE_PRIORITY = { gui: 0, face: 1, subscriber: 1, publisher: 2, smiley: 3, color: 3 };
  const nodeScore = svcMap => Object.keys(svcMap).reduce((min, svc) => {
    const p = SVC_NODE_PRIORITY[chaosBaseSvc(svc)] ?? 99; return p < min ? p : min;
  }, Infinity);

  function subdivide(zones, baseIcon, groupScope) {
    if (!zones.length) return [];
    const nodeGroups = new Map();
    for (const zone of zones) {
      for (const [svc, pods] of Object.entries(zone.pods || {})) {
        for (const pod of pods) {
          const n = pod.node || '__unknown__';
          if (!nodeGroups.has(n)) nodeGroups.set(n, {});
          (nodeGroups.get(n)[svc] = nodeGroups.get(n)[svc] || []).push(pod);
        }
      }
    }
    const keys = [...nodeGroups.keys()];
    const hasRealNodes = keys.some(k => k !== '__unknown__');
    const cardMod = groupScope === 'external' ? ' is-ew' : ' is-onprem';
    if (keys.length <= 1 || !hasRealNodes) {
      // Single card for the whole group → scope is the group itself
      return zones.map(z => ({ label: z.label, icon: z.icon, scope: groupScope, detail: null, cardMod, pods: z.pods }));
    }
    keys.sort((a, b) => {
      const sa = nodeScore(nodeGroups.get(a)), sb = nodeScore(nodeGroups.get(b));
      return sa !== sb ? sa - sb : a.localeCompare(b);
    });
    return keys.map((nodeName, i) => ({
      label: `Node ${i + 1}`, icon: baseIcon,
      scope: nodeName !== '__unknown__' ? 'node' : groupScope,
      detail: nodeName !== '__unknown__' ? nodeName : null,
      actualNode: nodeName !== '__unknown__' ? nodeName : null,
      cardMod,
      pods: nodeGroups.get(nodeName),
    }));
  }

  // EW-specific subdivision: tries node names first; falls back to per-IP cards
  // so each physical VM gets its own selectable card even when node labels are absent.
  function subdivideEW(zones) {
    if (!zones.length) return [];
    const cardMod = ' is-ew';

    // Pass 1: try grouping by pod.node
    const nodeGroups = new Map();
    for (const zone of zones) {
      for (const [svc, pods] of Object.entries(zone.pods || {})) {
        for (const pod of pods) {
          const n = pod.node || '__unknown__';
          if (!nodeGroups.has(n)) nodeGroups.set(n, {});
          (nodeGroups.get(n)[svc] = nodeGroups.get(n)[svc] || []).push(pod);
        }
      }
    }
    const nodeKeys = [...nodeGroups.keys()];
    const hasRealNodes = nodeKeys.some(k => k !== '__unknown__');

    if (hasRealNodes && nodeKeys.filter(k => k !== '__unknown__').length > 1) {
      // Multiple distinct nodes — one card per node
      return nodeKeys.filter(k => k !== '__unknown__').sort().map((nodeName, i) => ({
        label: `EW ${i + 1}`, icon: '🔗',
        scope: 'node', detail: nodeName, actualNode: nodeName, cardMod,
        pods: nodeGroups.get(nodeName),
      }));
    }

    // Pass 2: no reliable node names — group by pod IP (each EW endpoint has a unique IP)
    const ipGroups = new Map(); // ip/key → { _ip, _name, svc: [pod], … }
    for (const zone of zones) {
      for (const [svc, pods] of Object.entries(zone.pods || {})) {
        for (const pod of pods) {
          const key = pod.ip || pod.IP || pod.name || '__unknown__';
          if (!ipGroups.has(key)) ipGroups.set(key, { _ip: pod.ip || pod.IP || null, _name: pod.name || null });
          const entry = ipGroups.get(key);
          (entry[svc] = entry[svc] || []).push(pod);
        }
      }
    }

    if (ipGroups.size <= 1) {
      // Still only one group — single card for the whole EW group
      return zones.map(z => ({ label: z.label, icon: z.icon, scope: 'external', detail: null, cardMod, pods: z.pods }));
    }

    // One card per distinct EW IP — labeled by pod name or IP
    return [...ipGroups.entries()].map(([key, entry], i) => {
      const { _ip, _name, ...pods } = entry;
      const label = _name ? shortPodName(_name) : (_ip || `EW ${i + 1}`);
      return {
        label, icon: '🔗',
        scope: 'ewip', detail: _ip || key, actualNode: _name || _ip || key,
        cardMod, pods,
      };
    });
  }

  const cloud = infra.zones.filter(z => z.zone !== '').map(z => ({
    label: z.label, icon: z.icon, scope: 'zone', detail: z.zone, cardMod: '', pods: z.pods,
  }));
  const ew  = subdivideEW(infra.zones.filter(z => z.label === 'External Workload'));
  const op  = subdivide(infra.zones.filter(z => z.label === 'On-Premise'), '🖥️', 'onprem');

  // Per-node cards from cloud zones — requires K8s RBAC for node labels.
  // Lets operators target individual nodes within a zone (sits between Cloud and EWs).
  const cloudNodeGroups = new Map();
  for (const z of infra.zones.filter(z => z.zone !== '')) {
    for (const [svc, pods] of Object.entries(z.pods || {})) {
      for (const pod of pods) {
        const n = pod.node || '__unknown__';
        if (!cloudNodeGroups.has(n)) cloudNodeGroups.set(n, {});
        const g = cloudNodeGroups.get(n);
        (g[svc] = g[svc] || []).push(pod);
      }
    }
  }
  const cloudNodeKeys = [...cloudNodeGroups.keys()].filter(k => k !== '__unknown__');
  cloudNodeKeys.sort((a, b) => {
    const sa = nodeScore(cloudNodeGroups.get(a)), sb = nodeScore(cloudNodeGroups.get(b));
    return sa !== sb ? sa - sb : a.localeCompare(b);
  });
  const cloudNodes = cloudNodeKeys.map((nodeName, i) => ({
    label: `Node ${i + 1}`, icon: '🖥️',
    scope: 'node', detail: nodeName, actualNode: nodeName,
    cardMod: ' is-onprem', pods: cloudNodeGroups.get(nodeName),
  }));

  const groups = [];
  if (cloud.length)      groups.push({ icon: '🌐', title: 'Cloud',              cards: cloud });
  if (cloudNodes.length) groups.push({ icon: '🖥️', title: 'Nodes',              cards: cloudNodes });
  if (ew.length)         groups.push({ icon: '🔗', title: 'External Workloads', cards: ew });
  if (op.length)         groups.push({ icon: '🏢', title: 'On-Premise',         cards: op });
  return { groups, svcOrder };
}

function renderFITopology(infra, chaos, mode) {
  const root = document.getElementById('fi-topology');
  if (!root) return;

  if (!infra || !infra.zones || !infra.zones.length) {
    root.innerHTML = `<p class="fi-topo-empty">Topology unavailable (Kubernetes API not reachable). “All pods” targeting still works.</p>`;
    state._fiTopoSig = 'empty';
    return;
  }

  const { groups, svcOrder } = fiBuildTopologyGroups(infra);

  // Base services with >1 pod across the whole topology. For these, the service-level
  // aggregate (chaos[base]) is meaningless per-pod, so we must NOT fall back to it.
  const baseCounts = {};
  for (const z of infra.zones) {
    for (const [svcKey, pods] of Object.entries(z.pods || {})) {
      const b = chaosBaseSvc(svcKey);
      baseCounts[b] = (baseCounts[b] || 0) + (pods ? pods.length : 0);
    }
  }
  state._fiMultiBases = new Set(Object.keys(baseCounts).filter(b => baseCounts[b] > 1));

  // Signature to avoid repainting (and breaking hover/tooltips) every poll when nothing changed.
  const sig = JSON.stringify({
    s: state.fiScope, d: state.fiScopeDetail,
    g: groups.map(gr => gr.cards.map(c => ({ l: c.label, sc: c.scope, dt: c.detail,
      p: Object.entries(c.pods).map(([k, v]) => [k, v.map(p => [p.ip, p.smiley, p.smileyEdge, p.color, p.colorEdge,
        p.chaos && [p.chaos.errorFraction, p.chaos.latchFraction, p.chaos.maxRate, (p.chaos.delayBuckets || []).join('.'), p.chaos.latched]])]) }))),
    c: Object.fromEntries(Object.entries(chaos).map(([k, v]) => [k, v && [v.errorFraction, v.latchFraction, v.maxRate, (v.delayBuckets || []).join('.'), v.latched, v.available]])),
  });
  if (sig === state._fiTopoSig) return;
  state._fiTopoSig = sig;

  const isSel = (scope, detail) => state.fiScope === scope && (state.fiScopeDetail || null) === (detail || null);

  let html = '';

  for (const g of groups) {
    html += `<div class="infra-group"><div class="infra-group-header">
      <span class="infra-group-icon">${g.icon}</span><span class="infra-group-name">${esc(g.title)}</span>
    </div><div class="infra-zones-row">`;
    for (const c of g.cards) {
      const selCls  = isSel(c.scope, c.detail) ? ' fi-topo-selected' : '';
      const nodeTip = c.actualNode ? ` data-tooltip="${esc('Node: ' + c.actualNode)}"` : '';
      html += `<div class="infra-zone-card fi-topo-card${c.cardMod}${selCls}" data-scope="${c.scope}" data-detail="${esc(c.detail || '')}" role="button" tabindex="0">
        <div class="infra-zone-header">
          <span class="infra-zone-icon">${c.icon}</span>
          <span class="infra-zone-name"${nodeTip}>${esc(c.label)}</span>
          <span class="fi-topo-check">✓</span>
        </div>`;
      const seen = new Set();
      const renderSvc = (svc, pods) => {
        const base = chaosBaseSvc(svc);
        const meta = INFRA_SVC_META[base] || { icon: '⬡', label: svc };
        // Base services get their nice label ("Smiley"); instances keep their key ("smiley2")
        const dispLabel = INFRA_SVC_META[svc] ? INFRA_SVC_META[svc].label : svc;
        let h = `<div class="infra-svc-group"><div class="infra-svc-label">
          <span class="infra-svc-icon">${meta.icon}</span><span class="infra-svc-name">${esc(dispLabel)}</span>
        </div>`;
        for (const pod of pods) {
          const fiColorAttr = (base === 'color' && pod.color) ? ` data-pod-color="${esc(pod.color)}"` : '';
          const fiColorEdgeAttr = (base === 'color' && pod.colorEdge) ? ` data-pod-color-edge="${esc(pod.colorEdge)}"` : '';
          h += `<div class="infra-pod-row fi-pod-row" data-tooltip="${esc(infraPodTip(pod, base))}"${fiColorAttr}${fiColorEdgeAttr}>
            <span class="infra-pod-name">${esc(shortPodName(pod.name || pod.ip))}</span>
            <span class="fi-pod-badges">${fiPodFaultBadges(base, chaos, pod)}</span>
          </div>`;
        }
        return h + `</div>`;
      };
      for (const svc of svcOrder) {
        if (chaosBaseSvc(svc) === 'gui') continue;
        if (c.pods[svc]?.length) { seen.add(svc); html += renderSvc(svc, c.pods[svc]); }
      }
      for (const [svc, pods] of Object.entries(c.pods)) {
        if (chaosBaseSvc(svc) === 'gui') continue;
        if (!seen.has(svc) && pods?.length) html += renderSvc(svc, pods);
      }
      html += `</div>`; // card
    }
    html += `</div></div>`; // zones-row, group
  }

  root.innerHTML = html;

  // Wire selection — clicking a card toggles its scope (re-click → All pods)
  root.querySelectorAll('.fi-topo-card').forEach(card => {
    const act = () => {
      const scope = card.dataset.scope, detail = card.dataset.detail || null;
      if (isSel(scope, detail)) setFIScope('all', null);   // toggle off
      else setFIScope(scope, detail);
    };
    card.addEventListener('click', act);
    card.addEventListener('keydown', e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); act(); } });
  });
}

// Live "will target N pods across M services" preview.
function updateFIEffectPreview() {
  const el = document.getElementById('fi-effect-preview');
  if (!el) return;
  const services = state.fiSelectedServices || [];
  if (services.length === 0) { el.textContent = 'No services selected.'; return; }
  if (state.fiScope === 'all') {
    el.textContent = `Will target all pods of: ${services.join(', ')}.`;
    return;
  }
  const targets = resolveScopeTargets(state.fiScope, state.fiScopeDetail, services);
  const hit = services.filter(s => targets[s]?.length);
  const pods = hit.reduce((n, s) => n + targets[s].length, 0);
  const skipped = services.filter(s => !targets[s]?.length);
  let txt = `Will target ${pods} pod${pods !== 1 ? 's' : ''} across ${hit.length} service${hit.length !== 1 ? 's' : ''}`;
  if (skipped.length) txt += ` (no pods in scope: ${skipped.join(', ')})`;
  el.textContent = txt + '.';
}

// Read the global panel params into a chaos body.
function fiGlobalBody() {
  const panel = document.getElementById('fi-global-panel');
  const num = f => parseFloat(panel.querySelector(`[data-field="${f}"]`)?.value || '0');
  const delays = [...panel.querySelectorAll('.fi-delay-bubble.active')].map(b => Number(b.dataset.ms));
  return {
    errorFraction: Math.round(num('errorFraction')),
    latchFraction: Math.round(num('latchFraction')),
    maxRate: num('maxRate'),
    delayBuckets: delays,
  };
}

// Immediately refetch infrastructure (which carries per-pod chaos) so topology badges
// reflect the apply without waiting for the next scheduled poll.
function refreshFITopologyBadges() {
  pollInfra();
}

async function applyFIGlobal() {
  const services = state.fiSelectedServices || [];
  const statusEl = document.getElementById('fi-global-status');
  const btn = document.getElementById('fi-global-apply');
  if (services.length === 0) { if (statusEl) statusEl.textContent = 'Select at least one service.'; return; }
  btn.disabled = true; if (statusEl) statusEl.textContent = 'Applying…';
  const res = await applyGlobalChaos(fiGlobalBody(), services, state.fiScope, state.fiScopeDetail);
  refreshFITopologyBadges();
  if (statusEl) statusEl.textContent = (res.failed ? '✗ ' : '✓ ') + res.summary;
  setTimeout(() => { if (statusEl) statusEl.textContent = ''; }, 5000);
  btn.disabled = false;
  showToast(`⚡ Fault injection ${res.summary}`, res.failed ? 'error' : 'success');
}

async function resetSelectedFI() {
  const services = state.fiSelectedServices || [];
  const statusEl = document.getElementById('fi-global-status');
  if (services.length === 0) { if (statusEl) statusEl.textContent = 'Select at least one service.'; return; }
  const body = { errorFraction: 0, latchFraction: 0, maxRate: 0, delayBuckets: [], forceUnlatch: true };
  if (statusEl) statusEl.textContent = 'Resetting…';
  const res = await applyGlobalChaos(body, services, state.fiScope, state.fiScopeDetail);
  refreshFITopologyBadges();
  if (statusEl) statusEl.textContent = '✓ Reset — ' + res.summary;
  setTimeout(() => { if (statusEl) statusEl.textContent = ''; }, 5000);
  showToast(`⚡ Reset ${res.summary}`, 'success');
}

async function forceUnlatchSelectedFI() {
  const services = state.fiSelectedServices || [];
  const statusEl = document.getElementById('fi-global-status');
  if (services.length === 0) { if (statusEl) statusEl.textContent = 'Select at least one service.'; return; }
  const res = await applyGlobalChaos({ forceUnlatch: true }, services, state.fiScope, state.fiScopeDetail);
  refreshFITopologyBadges();
  if (statusEl) statusEl.textContent = '✓ Unlatched — ' + res.summary;
  setTimeout(() => { if (statusEl) statusEl.textContent = ''; }, 5000);
}

async function refreshFIPodSelectors() {
  const svcs = ['smiley', 'color'];
  for (const svc of svcs) {
    const ep = FI_PODS_ENDPOINT[svc];
    if (!ep) continue;
    try {
      const pods = await fetchJSON(ep).catch(() => []);
      fiPodCache[svc] = pods;
      // Re-render pod selectors on any visible card
      document.querySelectorAll(`.fi-card[data-svc="${svc}"]`).forEach(card => {
        renderFIPodSelector(card, svc, pods);
      });
    } catch (_) {}
  }
}

function renderFaultInjection(chaos, mode) {
  const container = document.getElementById('fi-cards');
  if (!container) return;

  const desc = document.getElementById('fi-page-desc');
  if (desc && FI_BANNERS[mode]) desc.innerHTML = FI_BANNERS[mode];

  state.lastChaos = chaos; // cache for topology repaint on scope clicks / infra poll

  const svcs = mode === 'classic' ? CHAOS_SERVICES_CLASSIC : CHAOS_SERVICES_PUBSUB;

  // Build (or rebuild on mode change) the scenario bar + global panel
  if (fiBuiltMode !== mode) {
    state.fiSelectedServices = [...svcs]; // default: all services selected
    renderFIScenarioBar(mode);
    buildFIGlobalPanel(mode);
    renderFIOutageButtons(mode);
    fiBuiltMode = mode;
  }

  // Status & Targeting topology (infra-style); updates live fault badges every poll
  renderFITopology(state.lastInfra, chaos, mode);
  updateFITargetLabel();
  updateFIEffectPreview();

  svcs.forEach(svc => {
    let card = container.querySelector(`.fi-card[data-svc="${svc}"]`);
    // Sliders mirror the selected pod's actual state when pods are selected;
    // null = selected pod unknown (infra poll hasn't caught up) → don't clobber.
    const st = fiCardSourceState(svc, chaos) || { ...(chaos[svc] || { available: false }), _noSliderSync: true };

    if (!card) {
      card = document.createElement('div');
      card.className = 'fi-card';
      card.dataset.svc = svc;
      card.innerHTML = buildFICardHTML(svc, mode);
      container.appendChild(card);
      wireFICard(card, svc);
      // Load pod list for HTTP-based endpoints
      const ep = FI_PODS_ENDPOINT[svc];
      if (ep && ep !== '__controls__') {
        fetchJSON(ep).then(pods => {
          fiPodCache[svc] = pods;
          renderFIPodSelector(card, svc, pods);
        }).catch(() => {});
      }
    } else {
      const effEl = card.querySelector('.fi-effect-text');
      if (effEl) effEl.textContent = FI_EFFECTS[svc]?.[mode] || '';
    }

    updateFICard(card, svc, st);
    card.classList.toggle('fi-card-unavailable', !st.available);

    // Publisher/subscriber: refresh pod selector from controls state (polled every 3s)
    if (FI_PODS_ENDPOINT[svc] === '__controls__') {
      const ctrl = state.lastControls?.[svc];
      if (ctrl?.pods?.length) {
        const pods = ctrl.pods.map(p => ({ ip: p.PodIP || p.podIP || '', name: p.PodName || p.podName || p.PodIP || p.podIP, zone: p.Zone || p.zone }));
        fiPodCache[svc] = pods;
        renderFIPodSelector(card, svc, pods);
      }
    }
  });

  container.querySelectorAll('.fi-card').forEach(card => {
    if (!svcs.includes(card.dataset.svc)) card.remove();
  });
}

function buildFICardHTML(svc, mode) {
  const label  = CHAOS_SERVICE_LABELS[svc] || svc;
  const icon   = FI_ICONS[svc] || '⬡';
  const effect = FI_EFFECTS[svc]?.[mode] || '';
  const hasPods = !!FI_PODS_ENDPOINT[svc];

  const delayBubbles = FI_DELAY_PRESETS.map(ms =>
    `<button class="fi-delay-bubble" data-ms="${ms}" type="button">${ms}ms</button>`
  ).join('');

  return `
    <div class="fi-card-head">
      <span class="fi-card-icon">${icon}</span>
      <span class="fi-card-name">${esc(label)}</span>
      <span class="chaos-latched-badge" style="display:none">⚠ LATCHED</span>
    </div>
    <p class="fi-effect-text">${esc(effect)}</p>

    <div class="fi-param-group">
      <div class="fi-param-row">
        <div class="fi-param-label-row">
          <span class="fi-param-label" data-tooltip="% of requests that return HTTP 500. 0 = disabled.">Error %</span>
          <span class="fi-slider-val" id="fi-err-val-${svc}">0%</span>
        </div>
        <input type="range" class="fi-slider" data-field="errorFraction" min="0" max="100" value="0">
      </div>
      <div class="fi-param-row">
        <div class="fi-param-label-row">
          <span class="fi-param-label" data-tooltip="When an error fires, this % chance latches the pod into 599 state for 30 s. Force Unlatch recovers immediately.">Latch %</span>
          <span class="fi-slider-val" id="fi-latch-val-${svc}">0%</span>
        </div>
        <input type="range" class="fi-slider" data-field="latchFraction" min="0" max="100" value="0">
      </div>
      <div class="fi-param-row">
        <div class="fi-param-label-row">
          <span class="fi-param-label" data-tooltip="Max requests per second. Requests above this return HTTP 429. 0 = unlimited.">Max RPS</span>
          <span class="fi-slider-val" id="fi-rps-val-${svc}">Off</span>
        </div>
        <input type="range" class="fi-slider" data-field="maxRate" min="0" max="20" step="0.5" value="0">
      </div>
      <div class="fi-param-row">
        <span class="fi-param-label" data-tooltip="One delay value is picked randomly per request. Select any combination. Active = highlighted.">Delay (ms)</span>
        <div class="fi-delay-bubbles" data-field="delayBuckets">
          ${delayBubbles}
        </div>
      </div>
    </div>

    ${hasPods ? `<div class="fi-pods-section">
      <span class="fi-pods-label">Target pods</span>
      <div class="fi-pods-row" id="fi-pods-${svc}">
        <span class="fi-pods-loading">Loading pods…</span>
      </div>
    </div>` : ''}

    <div class="fi-card-actions">
      <button class="fi-btn fi-apply-btn" type="button">Apply</button>
      <button class="fi-btn fi-reset-btn secondary" type="button">Reset All</button>
      <button class="fi-btn fi-unlatch-btn danger" type="button" style="display:none">Force Unlatch</button>
    </div>
    <div class="fi-card-status"></div>`;
}

function wireFICard(card, svc) {
  // Sliders — any user edit marks the card dirty so the poll doesn't clobber
  // the composed values before Apply (cleared on Apply/Reset success).
  card.querySelectorAll('.fi-slider').forEach(slider => {
    const field = slider.dataset.field;
    slider.addEventListener('input', () => {
      card.dataset.fiDirty = '1';
      updateFISliderVal(card, svc, field, slider.value);
    });
  });

  // Delay bubbles — toggle selection
  card.querySelectorAll('.fi-delay-bubble').forEach(btn => {
    btn.addEventListener('click', () => {
      card.dataset.fiDirty = '1';
      btn.classList.toggle('active');
    });
  });

  // Buttons
  card.querySelector('.fi-apply-btn').addEventListener('click',   () => applyFIChaos(card, svc));
  card.querySelector('.fi-reset-btn').addEventListener('click',   () => resetFIChaos(card, svc));
  card.querySelector('.fi-unlatch-btn').addEventListener('click', () => unlatchFIChaos(card, svc));
}

function updateFISliderVal(card, svc, field, value) {
  const num = parseFloat(value);
  if (field === 'errorFraction') {
    const el = document.getElementById(`fi-err-val-${svc}`);
    if (el) el.textContent = num === 0 ? 'Off' : `${num}%`;
  } else if (field === 'latchFraction') {
    const el = document.getElementById(`fi-latch-val-${svc}`);
    if (el) el.textContent = num === 0 ? 'Off' : `${num}%`;
  } else if (field === 'maxRate') {
    const el = document.getElementById(`fi-rps-val-${svc}`);
    if (el) el.textContent = num === 0 ? 'Off' : `${num} RPS`;
  }
}

function updateFICard(card, svc, state) {
  if (!state.available && state.error) {
    card.querySelector('.fi-card-status').textContent = '⚠ ' + state.error;
  }

  // Never clobber the sliders while the user has unapplied edits (dirty), while
  // a control is focused, or when the source state is unknown (_noSliderSync).
  const focused = card.querySelector(':focus');
  if (!focused && card.dataset.fiDirty !== '1' && !state._noSliderSync) {
    const setSliderFI = (field, val) => {
      const s = card.querySelector(`[data-field="${field}"]`);
      if (s && document.activeElement !== s) {
        s.value = val;
        updateFISliderVal(card, svc, field, val);
      }
    };
    setSliderFI('errorFraction', state.errorFraction || 0);
    setSliderFI('latchFraction', state.latchFraction || 0);
    setSliderFI('maxRate', state.maxRate || 0);

    // Update delay bubble active state
    const active = new Set(state.delayBuckets || []);
    card.querySelectorAll('.fi-delay-bubble').forEach(btn => {
      btn.classList.toggle('active', active.has(Number(btn.dataset.ms)));
    });
  }

  const isLatched = !!state.latched;
  card.querySelector('.chaos-latched-badge').style.display = isLatched ? '' : 'none';
  card.querySelector('.fi-unlatch-btn').style.display      = isLatched ? '' : 'none';
}

// Find a pod by IP anywhere in the cached infra payload (any zone, any service).
function fiFindInfraPod(ip) {
  const infra = state.lastInfra;
  if (!infra || !infra.zones) return null;
  for (const z of infra.zones) {
    for (const pods of Object.values(z.pods || {})) {
      for (const p of pods) {
        if ((p.ip || p.IP) === ip) return p;
      }
    }
  }
  return null;
}

// Serving-state indicator (emoji / colour swatch) for a pod IP, looked up from the
// cached infra payload — lets the FI pod selector show which pod is which at a glance.
function fiPodServingHTML(baseSvc, ip) {
  const p = fiFindInfraPod(ip);
  return p ? infraPodStateEl(p, baseSvc) : '';
}

// The chaos state an FI card should display: the FIRST selected pod's per-pod
// chaos when pods are selected (the service VIP aggregate would show some other
// pod's state), else the service aggregate. Returns null when the selected
// pod's state is unknown — callers must then leave the sliders alone.
function fiCardSourceState(svc, chaos) {
  const sel = fiSelectedPods[svc] || [];
  if (sel.length > 0) {
    const pod = fiFindInfraPod(sel[0]);
    return (pod && pod.chaos) ? { ...pod.chaos, available: true } : null;
  }
  return chaos[svc] || { available: false };
}

// Re-render FI pod-selector pills when their serving glyphs first arrive or change.
// The smiley/color/face pills otherwise render only once (at card creation), which is
// usually before the first /api/infrastructure poll completes — so without this their
// emoji/swatch never shows. Signature-guarded so unchanged pills aren't repainted
// (repainting would kill open tooltips). Selection state survives re-render because
// renderFIPodSelector re-reads fiSelectedPods.
function refreshFIPillGlyphs() {
  for (const [svc, pods] of Object.entries(fiPodCache)) {
    if (!pods || !pods.length) continue;
    const card = document.querySelector(`.fi-card[data-svc="${svc}"]`);
    if (!card) continue;
    const sig = pods.map(p => fiPodServingHTML(svc, p.ip || p.IP || '')).join('|');
    if (card.dataset.pillSig === sig) continue;
    card.dataset.pillSig = sig;
    renderFIPodSelector(card, svc, pods);
  }
}

function renderFIPodSelector(card, svc, pods) {
  const row = card.querySelector(`#fi-pods-${svc}`);
  if (!row) return;
  if (!pods || pods.length === 0) {
    row.innerHTML = `<span class="fi-pods-loading">No pods discovered</span>`;
    return;
  }
  // Stable display order so the pills don't reshuffle on poll. Publisher/subscriber pods
  // arrive from /api/controls in nondeterministic (headless-DNS) order each 3s poll.
  pods = [...pods].sort((a, b) => {
    const an = a.name || a.Name || a.ip || a.IP || '';
    const bn = b.name || b.Name || b.ip || b.IP || '';
    if (an !== bn) return an < bn ? -1 : 1;
    const ai = a.ip || a.IP || '', bi = b.ip || b.IP || '';
    return ai < bi ? -1 : ai > bi ? 1 : 0;
  });
  if (!fiSelectedPods[svc]) fiSelectedPods[svc] = [];

  row.innerHTML = pods.map(p => {
    const ip   = p.ip || p.IP || '';
    const name = shortPodName(p.name || p.Name || ip);
    const sel  = fiSelectedPods[svc].includes(ip);
    const stateHTML = fiPodServingHTML(svc, ip);
    const statePrefix = stateHTML ? `<span class="pod-pill-state">${stateHTML}</span>` : '';
    return `<button class="pod-pill${sel ? ' active' : ''}" data-ip="${esc(ip)}" data-tooltip="${esc(podTooltipText(normPod(p)))}">${statePrefix}${esc(name)}</button>`;
  }).join('');

  row.querySelectorAll('.pod-pill').forEach(btn => {
    btn.addEventListener('click', () => {
      const ip = btn.dataset.ip;
      if (!fiSelectedPods[svc]) fiSelectedPods[svc] = [];
      const idx = fiSelectedPods[svc].indexOf(ip);
      if (idx === -1) fiSelectedPods[svc].push(ip);
      else            fiSelectedPods[svc].splice(idx, 1);
      btn.classList.toggle('active', fiSelectedPods[svc].includes(ip));
      // Selection changed — immediately show the (now) targeted pod's actual
      // chaos state instead of waiting for the next poll. Fresh context, so any
      // unapplied edits are discarded on purpose.
      const st = fiCardSourceState(svc, state.lastChaos || {});
      if (st) {
        delete card.dataset.fiDirty;
        updateFICard(card, svc, st);
      }
    });
  });
}

async function applyFIChaos(card, svc) {
  const body = {};
  const ef = card.querySelector('[data-field="errorFraction"]');
  if (ef) body.errorFraction = parseInt(ef.value, 10);
  const lf = card.querySelector('[data-field="latchFraction"]');
  if (lf) body.latchFraction = parseInt(lf.value, 10);
  const mr = card.querySelector('[data-field="maxRate"]');
  if (mr) body.maxRate = parseFloat(mr.value) || 0;
  // Delay buckets from active bubbles
  const active = [...card.querySelectorAll('.fi-delay-bubble.active')].map(b => Number(b.dataset.ms));
  body.delayBuckets = active;

  // Per-pod targeting
  const sel = fiSelectedPods[svc] || [];
  if (sel.length > 0) body.pods = sel;

  const statusEl = card.querySelector('.fi-card-status');
  const btn = card.querySelector('.fi-apply-btn');
  btn.disabled = true; statusEl.textContent = 'Applying…';
  try {
    const r = await fetch(`/api/chaos/${svc}`, { method: 'PUT', headers: {'Content-Type':'application/json'}, body: JSON.stringify(body) });
    const d = await r.json();
    if (r.ok) {
      fiCacheAppliedChaos(svc, sel, body); // sliders survive polls until infra refetch lands
      delete card.dataset.fiDirty;
      refreshFITopologyBadges();  // refetch infra so per-pod badges reflect this card apply
      const ps = chaosPodStatus(d);
      statusEl.textContent = ps.text;
      if (!ps.isError) {
        setTimeout(() => { statusEl.textContent = ''; }, 3000);
        showToast(`⚡ ${CHAOS_SERVICE_LABELS[svc] || svc} fault injection applied`, 'success');
      } else {
        showToast(`⚠ ${CHAOS_SERVICE_LABELS[svc] || svc}: ${ps.text}`, 'warn');
      }
    } else {
      statusEl.textContent = '✗ ' + (d.error || r.statusText);
    }
  } catch (e) { statusEl.textContent = '✗ ' + e.message; }
  finally { btn.disabled = false; }
}

// Optimistically fold a just-applied chaos body into the cached per-pod state
// (state.lastInfra) and the service aggregate (state.lastChaos). Without this,
// poll ticks that land between the apply and the infra refetch would repaint
// the card's sliders with the pre-apply values.
function fiCacheAppliedChaos(svc, selectedIPs, body) {
  // Mirror the services' bucket sanitization: zeros survive only in mixed lists
  const buckets = (body.delayBuckets || []).filter(ms => ms >= 0);
  const applied = {
    errorFraction: body.errorFraction ?? 0,
    latchFraction: body.latchFraction ?? 0,
    maxRate:       body.maxRate ?? 0,
    delayBuckets:  buckets.some(ms => ms > 0) ? buckets : [],
    available:     true,
  };
  if (body.forceUnlatch) applied.latched = false;
  const ips = selectedIPs.length ? selectedIPs : (fiPodCache[svc] || []).map(p => p.ip || p.IP || '');
  ips.forEach(ip => {
    const p = fiFindInfraPod(ip);
    if (p) p.chaos = { ...(p.chaos || {}), ...applied };
  });
  if (!selectedIPs.length && state.lastChaos && state.lastChaos[svc]) {
    state.lastChaos[svc] = { ...state.lastChaos[svc], ...applied };
  }
}

async function resetFIChaos(card, svc) {
  const body = { errorFraction: 0, latchFraction: 0, delayBuckets: [], maxRate: 0, forceUnlatch: true };
  const sel = fiSelectedPods[svc] || [];
  if (sel.length > 0) body.pods = sel;
  const statusEl = card.querySelector('.fi-card-status');
  statusEl.textContent = 'Resetting…';
  try {
    const r = await fetch(`/api/chaos/${svc}`, { method: 'PUT', headers: {'Content-Type':'application/json'}, body: JSON.stringify(body) });
    const d = await r.json();
    if (r.ok) {
      fiCacheAppliedChaos(svc, sel, body);
      delete card.dataset.fiDirty;
      refreshFITopologyBadges();  // refetch infra so per-pod badges reflect this card reset
      const ps = chaosPodStatus(d);
      statusEl.textContent = ps.text.replace('Applied', 'Reset');
      if (!ps.isError) setTimeout(() => { statusEl.textContent = ''; }, 3000);
    } else { statusEl.textContent = '✗ ' + (d.error || r.statusText); }
  } catch (e) { statusEl.textContent = '✗ ' + e.message; }
}

async function unlatchFIChaos(card, svc) {
  const body = { forceUnlatch: true };
  const sel = fiSelectedPods[svc] || [];
  if (sel.length > 0) body.pods = sel;
  const statusEl = card.querySelector('.fi-card-status');
  statusEl.textContent = 'Unlatching…';
  try {
    const r = await fetch(`/api/chaos/${svc}`, { method: 'PUT', headers: {'Content-Type':'application/json'}, body: JSON.stringify(body) });
    if (r.ok) { statusEl.textContent = '✓ Unlatched'; setTimeout(() => { statusEl.textContent = ''; }, 3000); }
    else { const d = await r.json().catch(() => ({})); statusEl.textContent = '✗ ' + (d.error || r.statusText); }
  } catch (e) { statusEl.textContent = '✗ ' + e.message; }
}
