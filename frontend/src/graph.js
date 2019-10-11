import React from 'react'
import Chart from 'react-apexcharts'
import { Typography } from 'antd'
const { Title } = Typography

export default class Graph extends React.Component {
  render() {
    const queue = this.props.data.queue
    const timeline = this.props.data.timeline.map(entry => {
      return Object.assign({}, entry, { from: (new Date(entry.from * 1000)).toString() })
    })
    return (
      <div className="chart">
        <Title level={2}>Queue: {queue}</Title>
        <Chart
          options={{
            dataLabels: {
              enabled: false
            },
            stroke: {
              curve: 'smooth'
            },
            colors: ["#FF1654", "#17d352", "#ED9D12"],
            xaxis: {
              type: 'datetime',
              categories: timeline.map(el => el.from)
            }
          }}
          series={[
            {
              name: "Errors",
              data: timeline.map(el => el.errors ? el.errors.length : 0)
            },
            {
              name: "Success",
              data: timeline.map(el => el.successes || 0)
            },
            {
              name: "Retries",
              data: timeline.map(el => el.retries || 0)
            }
          ]}
          type="area"
          width="100%"
          height="300px"
        />
      </div>
    )
  }
}
