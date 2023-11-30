
const url = 'https://3xi97uokw5.execute-api.us-east-1.amazonaws.com/stage/mfl-scoring?output=json'

async function getScoring() {
    try {
        const response = await fetch(url)
        if (!response.ok) { throw new Error('Request Failed')}

        const data = await response.text()
        return data
    } catch (error) {
        console.log(error);
    }
}

console.log(getScoring());