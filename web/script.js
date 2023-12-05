const url = 'https://3xi97uokw5.execute-api.us-east-1.amazonaws.com/stage/mfl-scoring?output=json'

async function fetchScoring() {
    try {
        const response = await fetch(url)
        if (!response.ok) { throw new Error('Request Failed') }

        const data = await response.json()
        return data
    } catch (error) {
        console.log(error);
    }
}

async function logFetchScoring() {
  console.log(await fetchScoring());
}

async function displayTeams() {
    const results = await fetchScoring();
  
    results.forEach((team) => {
      const tr = document.createElement('tr');
      tr.classList.add('table-row');
      tr.innerHTML = `
      <tr>
        <td scope="col" class="table-data">${team.TeamName}</td>
        <td scope="col" class="table-data">${team.OwnerName.split(" ")[0]}</td>
        <td scope="col" class="table-data">${team.Record}</td>
        <td scope="col" class="table-data">${team.PointsFor}</td>
        <td scope="col" class="table-data">${parseFloat(team.PointScore)}</td>
        <td scope="col" class="table-data">${team.RecordScore}</td>
        <td scope="col" class="table-data"><b>${team.TotalScore}</b></td>
        <td scope="col" class="table-data">${team.AllPlayRecord}</td>
        <td scope="col" class="table-data">${team.AllPlayPercentage}</td>
      </tr>
      `

      document.querySelector("#table-content").appendChild(tr);
    });
}

function initialize() {
  displayTeams();
}

// Wait for the DOMContentLoaded event
document.addEventListener('DOMContentLoaded', initialize);