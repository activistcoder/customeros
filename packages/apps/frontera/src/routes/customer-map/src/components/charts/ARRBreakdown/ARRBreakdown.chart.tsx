import { PatternLines } from '@visx/pattern';
import {
  XYChart,
  Tooltip,
  BarStack,
  BarSeries,
  AnimatedGrid,
  AnimatedAxis,
} from '@visx/xychart';

import { cn } from '@ui/utils/cn';
import { formatCurrency } from '@spaces/utils/getFormattedCurrencyNumber';

import { mockData } from './mock';
import { Legend } from '../../Legend';
import { getMonthLabel } from '../util';

export type ARRBreakdownDatum = {
  month: number;
  upsells: number;
  churned: number;
  renewals: number;
  downgrades: number;
  cancellations: number;
  newlyContracted: number;
};

interface ARRBreakdownProps {
  width: number;
  height?: number;
  hasContracts?: boolean;
  data: ARRBreakdownDatum[];
}

const getX = (d: ARRBreakdownDatum) => getMonthLabel(d.month);

const ARRBreakdownChart = ({
  width,
  data: _data,
  hasContracts,
}: ARRBreakdownProps) => {
  const data = hasContracts ? _data : mockData;

  const colors = {
    gray700: '#344054',
    greenLight200: hasContracts ? '#D0F8AB' : '#F9FAFB',
    greenLight400: hasContracts ? '#85E13A' : '#D0D5DD',
    warning300: hasContracts ? '#FEC84B' : '#F2F4F7',
    warning600: hasContracts ? '#DC6803' : '#EAECF0',
    warning950: hasContracts ? '#4E1D09' : '#D0D5DD',
    greenLight700: hasContracts ? '#3B7C0F' : '#D0D5DD',
    greenLight500: hasContracts ? '#66C61C' : '#EAECF0',
  };

  const colorScale = {
    NewlyContracted: colors.greenLight700,
    Renewals: colors.greenLight500,
    Upsells: colors.greenLight200,
    Downgrades: colors.warning300,
    Cancellations: colors.warning600,
    Churned: colors.warning950,
  };

  const isMissingData = (dataPoint: keyof ARRBreakdownDatum) =>
    data.every((d) => d[dataPoint] === 0);

  const legendData = [
    {
      label: 'Newly contracted',
      color: colorScale.NewlyContracted,
      isMissingData: isMissingData('newlyContracted'),
    },
    {
      label: 'Renewals',
      color: colorScale.Renewals,
      isMissingData: isMissingData('renewals'),
    },
    {
      label: 'Upsells',
      color: colorScale.Upsells,
      borderColor: colors.greenLight400,
      isMissingData: isMissingData('upsells'),
    },
    {
      label: 'Downgrades',
      color: colorScale.Downgrades,
      borderColor: !hasContracts ? colors.greenLight400 : undefined,
      isMissingData: isMissingData('downgrades'),
    },
    {
      label: 'Cancellations',
      color: colorScale.Cancellations,
      isMissingData: isMissingData('cancellations'),
    },
    {
      label: 'Churned',
      color: colorScale.Churned,
      isMissingData: isMissingData('churned'),
    },
  ];

  const getBarColor = (key: keyof typeof colorScale, barIndex: number) =>
    barIndex === data.length - 1 ? `url(#stripes-${key})` : colorScale[key];

  return (
    <>
      <Legend data={legendData} />
      <XYChart
        height={200}
        width={width || 500}
        margin={{ top: 12, right: 0, bottom: 20, left: 0 }}
        xScale={{
          type: 'band',
          paddingInner: 0.4,
          paddingOuter: 0.4,
        }}
        yScale={{ type: 'linear' }}
      >
        {Object.entries(colorScale).map(([key, color]) => (
          <PatternLines
            key={key}
            id={`stripes-${key}`}
            height={8}
            width={8}
            stroke={color}
            strokeWidth={2}
            orientation={['diagonal']}
          />
        ))}
        <BarStack offset='diverging'>
          <BarSeries
            dataKey='Churned'
            data={data}
            xAccessor={(d) => getMonthLabel(d.month)}
            yAccessor={(d) => -d.churned}
            colorAccessor={(_, i) => getBarColor('Churned', i)}
          />
          <BarSeries
            dataKey='Cancelations'
            data={data}
            xAccessor={(d) => getMonthLabel(d.month)}
            yAccessor={(d) => -d.cancellations}
            colorAccessor={(_, i) => getBarColor('Cancellations', i)}
          />
          <BarSeries
            dataKey='Downgrades'
            data={data}
            radiusBottom
            radius={4}
            xAccessor={(d) => getMonthLabel(d.month)}
            yAccessor={(d) => -d.downgrades}
            colorAccessor={(_, i) => getBarColor('Downgrades', i)}
          />

          <BarSeries
            dataKey='Newly Contracted'
            data={data}
            xAccessor={(d) => getMonthLabel(d.month)}
            yAccessor={(d) => d.newlyContracted}
            colorAccessor={(_, i) => getBarColor('NewlyContracted', i)}
          />
          <BarSeries
            dataKey='Renewals'
            data={data}
            xAccessor={(d) => getMonthLabel(d.month)}
            yAccessor={(d) => d.renewals}
            colorAccessor={(_, i) => getBarColor('Renewals', i)}
          />
          <BarSeries
            dataKey='Upsells'
            data={data}
            radius={4}
            radiusTop
            xAccessor={(d) => getMonthLabel(d.month)}
            yAccessor={(d) => d.upsells}
            colorAccessor={(_, i) => getBarColor('Upsells', i)}
          />
        </BarStack>

        <AnimatedGrid
          columns={false}
          numTicks={1}
          lineStyle={{ stroke: 'white', strokeWidth: 2 }}
        />

        <AnimatedAxis
          orientation='bottom'
          hideAxisLine
          hideTicks
          tickLabelProps={{
            fontSize: 12,
            fontWeight: 'medium',
            fontFamily: `var(--font-barlow)`,
          }}
        />
        <Tooltip
          snapTooltipToDatumY
          snapTooltipToDatumX
          style={{
            position: 'absolute',
            padding: '8px 12px',
            background: colors.gray700,
            borderRadius: '8px',
          }}
          renderTooltip={({ tooltipData }) => {
            const xLabel = getX(
              tooltipData?.nearestDatum?.datum as ARRBreakdownDatum,
            );
            const values = tooltipData?.nearestDatum
              ?.datum as ARRBreakdownDatum;

            const sumPositives =
              values.newlyContracted + values.renewals + values.upsells;
            const sumNegatives =
              values.churned + values.cancellations + values.downgrades;

            const totalSum = sumPositives - sumNegatives;

            return (
              <div className='flex flex-col'>
                {hasContracts ? (
                  <>
                    <div className='flex justify-between items-center'>
                      <p className='text-white font-semibold text-sm'>
                        {xLabel}
                      </p>
                      <p className='text-white font-semibold text-sm'>
                        {formatCurrency(totalSum)}
                      </p>
                    </div>
                    <div className='flex flex-col'>
                      <TooltipEntry
                        label='Upsells'
                        value={values.upsells}
                        color={colorScale.Upsells}
                        isMissingData={isMissingData('upsells')}
                      />
                      <TooltipEntry
                        label='Renewals'
                        value={values.renewals}
                        color={colorScale.Renewals}
                        isMissingData={isMissingData('renewals')}
                      />
                      <TooltipEntry
                        label='Newly contracted'
                        value={values.newlyContracted}
                        color={colorScale.NewlyContracted}
                        isMissingData={isMissingData('newlyContracted')}
                      />
                      <TooltipEntry
                        label='Churned'
                        value={values.churned}
                        color={colorScale.Churned}
                        isMissingData={isMissingData('churned')}
                      />
                      <TooltipEntry
                        label='Cancellations'
                        value={values.cancellations}
                        color={colorScale.Cancellations}
                        isMissingData={isMissingData('cancellations')}
                      />
                      <TooltipEntry
                        label='Downgrades'
                        value={values.downgrades}
                        color={colorScale.Downgrades}
                        isMissingData={isMissingData('downgrades')}
                      />
                    </div>
                  </>
                ) : (
                  <p className='text-white font-semibold text-sm'>
                    No data yet
                  </p>
                )}
              </div>
            );
          }}
        />
      </XYChart>
      <p className='text-gray-500 text-xs mt-2'>
        <i>*Key data missing.</i>
      </p>
    </>
  );
};

const TooltipEntry = ({
  color,
  label,
  value,
  isMissingData,
}: {
  color: string;
  label: string;
  value: number;
  isMissingData?: boolean;
}) => {
  return (
    <div className='flex items-center gap-4'>
      <div className='flex items-center flex-1 gap-2'>
        <div
          className='flex w-2 h-2 rounded-full border border-white'
          style={{ backgroundColor: color }}
        />
        <p className='text-white text-sm'>{label}</p>
      </div>
      <div className='flex'>
        <p
          className={cn(
            isMissingData ? 'text-gray-400' : 'text-white',
            'text-sm',
          )}
        >
          {isMissingData ? '*' : formatCurrency(value)}
        </p>
      </div>
    </div>
  );
};

export default ARRBreakdownChart;