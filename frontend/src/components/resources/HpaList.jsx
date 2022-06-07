import React from 'react'

export const HpaList = ({ data }) => {
    return (
        <div className='bg-gray-500 rounded-sm shadow-md'>
            <table className='min-w-full'>
                <thead className='border-b border-gray-700'>
                    <th className="text-sm font-medium text-gray-900 px-6 py-4 text-left">Name</th>
                    <th className="text-sm font-medium text-gray-900 px-6 py-4 text-left">Namespace</th>
                    <th className="text-sm font-medium text-gray-900 px-6 py-4 text-left">Min Pods</th>
                    <th className="text-sm font-medium text-gray-900 px-6 py-4 text-left">Max Pods</th>
                    <th className="text-sm font-medium text-gray-900 px-6 py-4 text-left">Replicas</th>
                    <th className="text-sm font-medium text-gray-900 px-6 py-4 text-left">Status</th>
                </thead>
                <tbody>
                    {data.map(el => (
                        <tr className='border-b border-gray-700'>
                            <td className='text-sm text-gray-900 font-light px-6 py-4 whitespace-nowrap'>{el.metadata.name}</td>
                            <td className='text-sm text-gray-900 font-light px-6 py-4 whitespace-nowrap'>{el.metadata.namespace}</td>
                            <td className='text-sm text-gray-900 font-light px-6 py-4 whitespace-nowrap'>{el.spec.minReplicas}</td>
                            <td className='text-sm text-gray-900 font-light px-6 py-4 whitespace-nowrap'>{el.spec.maxReplicas}</td>
                            <td className='text-sm text-gray-900 font-light px-6 py-4 whitespace-nowrap'>{el.status.currentReplicas}</td>
                            <td className='text-sm text-gray-900 font-light px-6 py-4 whitespace-nowrap'>Able to scale</td>

                        </tr>
                    ))
                    }

                </tbody>
            </table>

        </div>
    )
}