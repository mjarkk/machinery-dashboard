import React from 'react'
import './App.css'
import { get } from './logic/calles'
import Graph from './graph'
// import { DatePicker } from 'antd'

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      data: [],
    }
    this.setup()
  }
  async setup() {
    try {
      const res = await get('')
      const { data } = res
      this.setState({ data })
    } catch (error) {
      console.log("Can't fetch data network, Error:", error)
    }
    setTimeout(() => this.setup(), 10000)
  }
  render() {
    return (
      <div className="main">
        {this.state.data.map((el, id) =>
          <Graph key={id} data={el} />
        )}
      </div>
    )
  }
}

export default App;
