import React from 'react'
import Chart from 'react-apexcharts'
import { Typography } from 'antd'
const { Title } = Typography

export default class Graph extends React.Component {
  render() {
    const queue = this.props.data.queue
    const timeline = this.props.data.timeline.map(entry => {
      entry.From = (new Date(entry.From * 1000)).toString()
      return entry
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
            colors: ["#FF1654", "#17d352"],
            xaxis: {
              type: 'datetime',
              categories: timeline.map(el => el.From)
            }
          }}
          series={[
            {
              name: "Errors",
              data: timeline.map(el => el.Errors.length)
            }, {
              name: "Success",
              data: timeline.map(el => el.Successes)
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
