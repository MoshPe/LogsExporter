const name = 'ServiceB'

let fullName = ''
async function main() {
    for (let i = 1; i <= 26; i++) {
        let charCode = i + 64;
        let letter = String.fromCharCode(charCode);
        for (let j = 1; j <= 26; j++) {
            let charCodeB = j + 64;
            let letterB = String.fromCharCode(charCodeB);
            await Sleep(100)
            fullName += `${name}_${letter}_${letterB}, ${String(Date.now())},`
        }
    }
    console.log(fullName)
}


function Sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}
main()
